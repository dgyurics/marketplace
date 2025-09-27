package types

import "errors"

var ErrNotFound = errors.New("resource not found")
var ErrUniqueConstraintViolation = errors.New("unique constraint violation")
var ErrConstraintViolation = errors.New("constraint violation")
