// Package-level sentinel errors for consistent error handling across services.
// Use errors.Is(err, types.ErrXxx) to check without coupling to message strings.
package types

import "errors"

var ErrNotFound = errors.New("resource not found")
var ErrUniqueConstraintViolation = errors.New("unique constraint violation")
var ErrConstraintViolation = errors.New("constraint violation")
var ErrInvalidInput = errors.New("invalid input")
