package middleware

import (
	"net/http"
)

const MaxBodyBytes = int64(65536) // 64 KB

// LimitBodySizeMiddleware limits the size of the request body.
func LimitBodySizeMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
			next.ServeHTTP(w, r)
		})
	}
}
