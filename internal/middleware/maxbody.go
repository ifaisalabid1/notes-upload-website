package middleware

import (
	"net/http"
	"strings"
)

const defaultMaxBodyBytes = 1 << 20

func MaxBodySize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "multipart/form-data") {
			next.ServeHTTP(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, defaultMaxBodyBytes)
		next.ServeHTTP(w, r)
	})
}
