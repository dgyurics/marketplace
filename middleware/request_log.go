package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// RequestLoggerMiddleware logs HTTP requests using slog.
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		// Structured logging with slog
		slog.Info("HTTP Request",
			slog.String("method", r.Method),
			slog.String("path", r.RequestURI),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.Duration("duration", duration),
		)
	})
}
