package types

import "errors"

// HTTPError represents an error that occurred during an HTTP request.
type HTTPError struct {
	Message    string
	StatusCode int
	err        error
}

func (h HTTPError) Error() string {
	return ""
}

func NewAPIError(statusCode int, message string, err error) HTTPError {
	return HTTPError{
		StatusCode: statusCode,
		Message:    message,
		err:        err,
	}
}

var ErrNotFound = errors.New("resource not found")
var ErrInvalidRequest = errors.New("invalid request")
var ErrUniqueConstraintViolation = errors.New("unique constraint violation")

type DatabaseError string

const (
	UniqueConstraintViolation DatabaseError = "UniqueConstraintViolation"
	NotNullViolation          DatabaseError = "NotNullViolation"
	CheckConstraintViolation  DatabaseError = "CheckConstraintViolation"
	ForeignKeyViolation       DatabaseError = "ForeignKeyViolation"
	UnknownDatabaseError      DatabaseError = "UnknownDatabaseError"
)
