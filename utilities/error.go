package utilities

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/lib/pq"
)

// ConvertToDatabaseError converts a raw error [err] to a [models.DatabaseError].
// Intended to be used in the service layer to convert database errors to an [models.HTTPError].
func ConvertToDatabaseError(err error) models.DatabaseError {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return models.UniqueConstraintViolation
		case "23502":
			return models.NotNullViolation
		case "23514":
			return models.CheckConstraintViolation
		case "23503":
			return models.ForeignKeyViolation
		default:
			return models.UnknownDatabaseError
		}
	}
	return models.UnknownDatabaseError
}

// RespondWithError logs the error and responds with a generic error message
// Exposing the actual error message to the client can be a security risk
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
	case http.StatusInternalServerError:
		message = "something went wrong"
	}
	http.Error(w, message, code)
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

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
