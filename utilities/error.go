package utilities

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// RespondWithError logs the error and responds with a generic error message
// Exposing the actual error to the client can be a security risk
func RespondWithError(w http.ResponseWriter, r *http.Request, code int, message string) {
	slog.Error("request_error",
		"status_code", code,
		"message", message,
		"method", r.Method,
		"uri", r.RequestURI,
		"ip", getIPAddress(r),
		"user_agent", r.UserAgent(),
		"referer", r.Referer(),
		"request_body", r.Body,
	)
	// Respond with a generic error message
	switch code {
	case http.StatusNotFound:
		message = "resource not found"
	case http.StatusBadRequest:
		message = "bad request"
	case http.StatusInternalServerError:
		message = "something went wrong"
	}
	http.Error(w, message, code)
}

// RespondWithJSON responds with a JSON payload, setting the appropriate headers and status code
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// RespondSuccess responds with a 200 OK status code
func RespondSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
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
