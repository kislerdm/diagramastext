package handler

import "net/http"

func ResponseTypeHandler(mimeType string, next http.Handler) http.Handler {
	return responseTypeHandler{
		t:    mimeType,
		next: next,
	}
}

type responseTypeHandler struct {
	t    string
	next http.Handler
}

func (h responseTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", h.t)
	if h.next != nil {
		h.next.ServeHTTP(w, r)
	}
}
