package types

// HTTPError represents an error that occurred during an HTTP request.
type HTTPError struct {
	Message    string
	StatusCode int
	Error      error
}

func NewAPIError(statusCode int, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		Error:      err,
	}
}

type DatabaseError string

const (
	UniqueConstraintViolation DatabaseError = "UniqueConstraintViolation"
	NotNullViolation          DatabaseError = "NotNullViolation"
	CheckConstraintViolation  DatabaseError = "CheckConstraintViolation"
	ForeignKeyViolation       DatabaseError = "ForeignKeyViolation"
	UnknownDatabaseError      DatabaseError = "UnknownDatabaseError"
)
