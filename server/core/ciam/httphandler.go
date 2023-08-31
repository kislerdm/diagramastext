// Package ciam to authN/Z users
package ciam

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

// HTTPHandler initializes the CIAM client.
func HTTPHandler(
	clientRepository RepositoryCIAM, clientEmail SMTPClient, privateKey ed25519.PrivateKey,
) (func(next http.Handler) http.Handler, error) {
	if clientRepository == nil {
		return nil, errors.New("repo client is required")
	}
	if clientEmail == nil {
		return nil, errors.New("email client is required")
	}
	issuer, err := NewIssuer(privateKey)
	if err != nil {
		return nil, err
	}
	return func(next http.Handler) http.Handler {
		return client{
			clientRepository: clientRepository,
			clientEmail:      clientEmail,
			tokenIssuer:      issuer,
			logger:           log.New(os.Stderr, "", log.Lmicroseconds|log.LUTC|log.Lshortfile),
			next:             next,
		}
	}, nil
}

type client struct {
	next http.Handler

	logger           *log.Logger
	clientRepository RepositoryCIAM
	clientEmail      SMTPClient
	tokenIssuer      Issuer
}

func (c client) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/auth") && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":"` + r.Method + ` is not allowed"}`))
		return
	}

	switch p := r.URL.Path; p {
	case "/auth/anonym":
		c.signinAnonym(w, r)
		return
	case "/auth/init":
		c.signinUserInit(w, r)
		return
	case "/auth/confirm":
		c.signinUserInitSecretConfirmation(w, r)
		return
	case "/auth/refresh":
		c.refreshAccessToken(w, r)
		return
	default:
		user, found, err := c.readUserFromHeader(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"internal error"}`))
			c.logger.Println(err)
			return
		}

		if !found {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"no authentication token provided"}`))
			return
		}

		if p == "/quotas" {
			c.getQuotaUsage(w, r, user)
			return
		}

		if ok := c.validateRequestsQuotaUsage(w, r, user); !ok {
			return
		}

		r = r.WithContext(NewContext(r.Context(), user))

		if c.next != nil {
			c.next.ServeHTTP(w, r)
		}
	}
}

// getQuotaUsage reads current usage of the quota.
func (c client) getQuotaUsage(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":"` + r.Method + ` is not allowed"}`))
		return
	}

	quotas, err := getQuotaUsage(r.Context(), c.clientRepository, user)
	if err != nil {
		c.internalError(w, err)
		return
	}

	o, err := json.Marshal(quotas)
	if err != nil {
		c.internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(o)
	return
}

// checks if the requests' quota was exceeded.
func (c client) validateRequestsQuotaUsage(w http.ResponseWriter, r *http.Request, user *User) bool {
	quotasUsage, err := getQuotaUsage(r.Context(), c.clientRepository, user)
	if err != nil {
		c.internalError(w, err)
		return false
	}

	if quotasUsage.RateDay.Used >= quotasUsage.RateDay.Limit {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"quota exceeded"}`))
		c.logger.Printf("quota exceeded for user %s", user.ID)
		return false
	}

	if quotasUsage.RateMinute.Used >= quotasUsage.RateMinute.Limit {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"throttling quota exceeded"}`))
		c.logger.Printf("throttling quota exceeded for user %s", user.ID)
		return false
	}

	return true
}

// anonym's authentication flow:
//
//	Fingerprint found in DB -> No  -> Create \
//							   -> Yes ->  --	-> Generate id, refresh and access JWT -> Return generated tokens.
func (c client) signinAnonym(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	var req struct {
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"request parsing error"}`))
		c.logger.Println(err)
		return
	}
	if f, _ := regexp.MatchString(`^[a-f0-9]{40}$`, req.Fingerprint); !f {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"invalid request"}`))
		c.logger.Printf("%s is invalid fingerprint\n", req.Fingerprint)
		return
	}

	userID, isActive, err := c.clientRepository.LookupUserByFingerprint(r.Context(), req.Fingerprint)
	if err != nil {
		c.internalError(w, err)
		return
	}

	if userID != "" && !isActive {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"user was deactivated"}`))
		c.logger.Printf("user %s was deactivated\n", userID)
		return
	}

	if userID == "" {
		userID = utils.NewUUID()
		role := uint8(RoleAnonymUser)
		if err := c.clientRepository.CreateUser(
			r.Context(), userID, "", req.Fingerprint, true, &role,
		); err != nil {
			c.internalError(w, err)
			return
		}
	}

	o, err := c.issueTokens(
		r.Context(), User{ID: userID, Role: RoleAnonymUser}, "", req.Fingerprint,
	)
	if err != nil {
		c.internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(o)
}

// signinUserInit executes user's authentication flow:
//
//	Email found in DB -> No  -> Create \
//			 	   	  -> Yes ->	--	  -> Generate secret and id JWT -> Send secret to email -> Return id JWT
func (c client) signinUserInit(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	var req struct {
		Email       string `json:"email"`
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"request parsing error"}`))
		c.logger.Println(err)
		return
	}
	if req.Email == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"email must be provided"}`))
		return
	}

	const defaultExpirationSecret = 10 * time.Minute

	userID, isActive, err := c.clientRepository.LookupUserByEmail(r.Context(), req.Email)
	if err != nil {
		c.internalError(w, err)
		return
	}

	if userID == "" {
		userID = utils.NewUUID()
		role := uint8(RoleRegisteredUser)
		if err := c.clientRepository.CreateUser(
			r.Context(), userID, req.Email, req.Fingerprint, false, &role,
		); err != nil {
			c.internalError(w, err)
			return
		}
	} else if !isActive {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"user was deactivated"}`))
		c.logger.Printf("user %s was deactivated\n", userID)
		return
	}

	secret := generateOnetimeSecret()
	iat := time.Now().UTC()

	if err := c.clientRepository.WriteOneTimeSecret(r.Context(), userID, secret, iat); err != nil {
		c.internalError(w, err)
		return
	}

	if err := c.clientEmail.SendSignInEmail(req.Email, secret); err != nil {
		c.internalError(w, err)
		return
	}

	tkn, err := c.tokenIssuer.NewIDToken(
		userID, req.Email, req.Fingerprint, WithCustomIat(iat), WithValidityDuration(defaultExpirationSecret),
	)
	if err != nil {
		c.internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(tkn))
	return
}

func (c client) internalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{"error":"internal error"}`))
	c.logger.Println(err)
}

// signinUserInit executes user's authentication flow:
// compare secret provided by user against the reference.
func (c client) signinUserInitSecretConfirmation(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	var req struct {
		Token  string `json:"id_token"`
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"request parsing error"}`))
		c.logger.Println(err)
		return
	}
	if req.Token == "" || req.Secret == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"token and secret must be provided"}`))
		return
	}
	userID, email, fingerprint, err := c.tokenIssuer.ParseIDToken(req.Token)
	if err != nil {
		c.internalError(w, err)
		return
	}

	found, secretRef, _, err := c.clientRepository.ReadOneTimeSecret(r.Context(), userID)
	if err != nil {
		c.internalError(w, err)
		return
	}

	if !found {
		c.internalError(w, errors.New("no secret was sent"))
		return
	}

	if req.Secret != secretRef {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"secret is wrong"}`))
		return
	}

	if err := c.clientRepository.UpdateUserSetActive(r.Context(), userID); err != nil {
		c.internalError(w, err)
		return
	}

	_ = c.clientRepository.DeleteOneTimeSecret(r.Context(), userID)

	o, err := c.issueTokens(
		r.Context(), User{ID: userID, Role: RoleRegisteredUser}, email, fingerprint,
	)
	if err != nil {
		c.internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(o)
}

func (c client) issueTokens(_ context.Context, user User, email, fingerprint string) (
	[]byte, error,
) {
	iat := time.Now().UTC()

	idToken, err := c.tokenIssuer.NewIDToken(user.ID, email, fingerprint, WithCustomIat(iat))
	if err != nil {
		return nil, err
	}

	accessToken, err := c.tokenIssuer.NewAccessToken(user, WithCustomIat(iat))
	if err != nil {
		return nil, err
	}

	refreshToken, err := c.tokenIssuer.NewRefreshToken(user.ID, WithCustomIat(iat))
	if err != nil {
		return nil, err
	}

	return []byte(`{"id":"` + idToken + `","access":"` + accessToken + `","refresh":"` + refreshToken + `"}`), nil
}

func (c client) ParseAccessToken(_ context.Context, token string) (User, error) {
	return c.tokenIssuer.ParseAccessToken(token)
}

func (c client) refreshAccessToken(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	var req struct {
		Token string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"request parsing error"}`))
		c.logger.Println(err)
		return
	}
	if req.Token == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"token must be provided"}`))
		return
	}

	userID, err := c.tokenIssuer.ParseRefreshToken(req.Token)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"token is not valid"}`))
		c.logger.Println(err)
		return
	}

	found, isActive, roleID, email, fingerprint, err := c.clientRepository.ReadUser(r.Context(), userID)
	if err != nil {
		c.internalError(w, err)
		return
	}
	if !found {
		c.internalError(w, errors.New("user not found"))
		return
	}
	if !isActive {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"user was deactivated"}`))
		c.logger.Printf("user %s was deactivated\n", userID)
		return
	}

	iat := time.Now().UTC()

	accToken, err := c.tokenIssuer.NewAccessToken(User{ID: userID, Role: Role(roleID)}, WithCustomIat(iat))
	if err != nil {
		c.internalError(w, err)
		return
	}

	idToken, err := c.tokenIssuer.NewIDToken(userID, email, fingerprint, WithCustomIat(iat))
	if err != nil {
		c.internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"id":"` + idToken + `","access":"` + accToken + `"}`))
}

func (c client) readUserFromHeader(r *http.Request) (*User, bool, error) {
	key, found := readAuthHeaderValue(r.Header)

	if !found {
		return c.readUserFromApiKey(r)
	}

	user, err := c.tokenIssuer.ParseAccessToken(key)
	if err != nil {
		return nil, false, err
	}

	return &user, true, nil
}

func (c client) readUserFromApiKey(r *http.Request) (*User, bool, error) {
	key, found := readAPIKey(r.Header)
	if !found {
		return nil, false, nil
	}

	userID, err := c.clientRepository.GetActiveUserIDByActiveTokenID(r.Context(), key)
	if err != nil {
		return nil, false, err
	}

	found, isActive, roleID, _, _, err := c.clientRepository.ReadUser(r.Context(), userID)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, errors.New("user database integrity problem")
	}
	if !isActive {
		return nil, false, errors.New("user " + userID + " is deactivated")
	}

	return &User{
		ID:       userID,
		APIToken: key,
		Role:     Role(roleID),
	}, true, nil
}

func generateOnetimeSecret() string {
	const (
		charset = "0123456789abcdef"
		length  = 6
	)
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var b = make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GenerateCertificate() ed25519.PrivateKey {
	_, o, _ := ed25519.GenerateKey(rand.New(rand.NewSource(0)))
	return o
}
