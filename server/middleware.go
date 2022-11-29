package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func setResponseMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml;charset=utf-8")
		w.Header().Set("Allow", "GET, HEAD, POST, SUBSCRIBE, UNSUBSCRIBE")
		value := context.WithValue(r.Context(), "ip", r.RemoteAddr)
		h.ServeHTTP(w, r.WithContext(value))
	})
}

func whiteIPsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果白名单为空则允许所有
		if len(WhiteIPs) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		ip := strings.Split(r.RemoteAddr, ":")[0]
		if isContain(WhiteIPs, ip) {
			h.ServeHTTP(w, r)
			return
		}
		fmt.Fprint(w, "Permission denied")
		return
	})
}

func isContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}
