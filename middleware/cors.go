package middleware

import (
	"net/http"
	"strings"

	"github.com/dgyurics/marketplace/types"
)

func CORSMiddleware(config types.CORSConfig) func(http.Handler) http.Handler {
	allowedMethods := strings.Join(config.AllowedMethods, ", ")
	allowedHeaders := strings.Join(config.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Set CORS headers
			if origin != "" {
				// If specific origins are set, check if the request origin is allowed
				if len(config.AllowedOrigins) > 0 && config.AllowedOrigins[0] != "*" {
					allowed := false
					for _, o := range config.AllowedOrigins {
						if o == origin {
							allowed = true
							break
						}
					}
					if !allowed {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
