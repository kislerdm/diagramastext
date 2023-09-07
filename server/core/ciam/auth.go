package ciam

import (
	"net/http"
	"strings"
)

func readAuthHeaderValue(header http.Header) (key string, found bool) {
	const authorizationHeaderName = "Authorization"
	authHeader := header.Get(authorizationHeaderName)
	if authHeader == "" {
		authHeader = header.Get(strings.ToLower(authorizationHeaderName))
	}
	_, v, found := strings.Cut(authHeader, "Bearer ")
	return v, found
}

func readAPIKey(header http.Header) (key string, found bool) {
	const authorizationHeaderName = "X-API-KEY"
	key = header.Get(authorizationHeaderName)
	if key == "" {
		key = header.Get(strings.ToLower(authorizationHeaderName))
	}
	if key != "" {
		found = true
	}
	return
}
