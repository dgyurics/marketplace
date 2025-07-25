package middleware

import (
	"net/http"
	"strings"
)

const MaxBodyBytes = int64(65536) // 64 KB

// LimitBodySizeMiddleware limits the size of the request body.
func LimitBodySizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip size limit for image upload endpoints
		if strings.HasPrefix(r.URL.Path, "/images/") {
			next.ServeHTTP(w, r)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		next.ServeHTTP(w, r)
	})
}
