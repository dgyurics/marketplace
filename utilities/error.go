package utilities

import (
	"errors"

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
