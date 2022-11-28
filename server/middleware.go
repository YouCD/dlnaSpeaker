package server

import (
	"context"
	"net/http"
)

func setResponseMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml;charset=utf-8")
		w.Header().Set("Allow", "GET, HEAD, POST, SUBSCRIBE, UNSUBSCRIBE")
		value := context.WithValue(r.Context(), "ip", r.RemoteAddr)
		h.ServeHTTP(w, r.WithContext(value))
	})
}
