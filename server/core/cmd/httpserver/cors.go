package main

import "net/http"

type cordHandler struct {
	m    map[string]string
	next http.Handler
}

func (c cordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, v := range c.m {
		w.Header().Set(k, v)
		if k == "Access-Control-Allow-Origin" && (v == "" || v == "'*'") {
			w.Header().Set(k, "*")
		}
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if c.next != nil {
		c.next.ServeHTTP(w, r)
	}
}

func CORSHandler(corsHeader map[string]string, next http.Handler) http.Handler {
	return cordHandler{
		m:    corsHeader,
		next: next,
	}
}
