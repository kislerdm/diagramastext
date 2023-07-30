package awscognito

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

type Client struct {
	cognitoClient *cognitoidentityprovider.CognitoIdentityProvider
	appClientID   string
}

// NewCognitoClient initiate the cognito Client.
func NewCognitoClient(ctx context.Context, cognitoRegion string, cognitoAppClientID string) (*Client, error) {
	conf := &aws.Config{Region: aws.String(cognitoRegion)}

	sess, err := session.NewSession(conf)
	if err != nil {
		return nil, err
	}

	client := cognitoidentityprovider.New(sess)

	return &Client{
		cognitoClient: client,
		appClientID:   cognitoAppClientID,
	}, nil
}

// ParseAndValidateToken validates JWT.
func (c *Client) ParseAndValidateToken(ctx context.Context, token string) (jwt.Token, error) {
	// Todo:- MC remove the hardcoded strings
	publicKeyUrl := "https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json"
	formattedURL := fmt.Sprintf(publicKeyUrl, "string1", "string2")
	keySet, err := jwk.Fetch(ctx, formattedURL)
	if err != nil {
		return nil, err
	}

	retToken, err := jwt.Parse(
		[]byte(token),
		jwt.WithKeySet(keySet),
		jwt.WithValidate(true),
	)
	if err != nil {
		return nil, err
	}

	// Todo:- MC Continue on Verify the claims sec: "https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-verifying-a-jwt.html"

	// Todo:- MC Fix return type
	return retToken, nil
}
