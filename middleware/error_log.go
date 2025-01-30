package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
)

// LoggingMiddleware logs 500 errors and captures additional request details.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture request details before processing
		requestBody := captureRequestBody(r)

		// Wrap the ResponseWriter to capture the status code and response body
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code
			body:           bytes.NewBuffer(nil),
		}

		// Call the next handler in the chain
		next.ServeHTTP(lrw, r)

		// Determine the log level based on status code
		if lrw.statusCode >= 500 {
			slog.Error("request_error",
				"status_code", lrw.statusCode,
				"method", r.Method,
				"uri", r.RequestURI,
				"ip", getIPAddress(r),
				"user_agent", r.UserAgent(),
				"referer", r.Referer(),
				"request_body", requestBody,
				"response_message", lrw.body.String(),
			)
		} else if lrw.statusCode >= 400 {
			slog.Warn("request_warning",
				"status_code", lrw.statusCode,
				"method", r.Method,
				"uri", r.RequestURI,
				"ip", getIPAddress(r),
				"user_agent", r.UserAgent(),
				"referer", r.Referer(),
				"request_body", requestBody,
			)
		}
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code and response body
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// WriteHeader captures the status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body content
func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	lrw.body.Write(data)
	return lrw.ResponseWriter.Write(data)
}

// captureRequestBody extracts the request body for logging (only for safe methods)
func captureRequestBody(r *http.Request) string {
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			// Restore the request body so it can be read again
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			return string(bodyBytes)
		}
	}
	return ""
}

// getIPAddress extracts the client's IP address from request headers or remote address
func getIPAddress(r *http.Request) string {
	// Check the common headers used for proxies
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}
