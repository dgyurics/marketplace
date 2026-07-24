// Package-level sentinel errors for consistent error handling across services.
// Use errors.Is(err, types.ErrXxx) to check without coupling to message strings.
package types

import "errors"

var (
	ErrNotFound                  = errors.New("resource not found")
	ErrUniqueConstraintViolation = errors.New("unique constraint violation")
	ErrConstraintViolation       = errors.New("constraint violation")
	ErrInvalidInput              = errors.New("invalid input")
)

type InsufficientStockItem struct {
	Product   Product `json:"product"`
	Quantity  int     `json:"quantity"`
	Inventory int     `json:"inventory"`
}

type InsufficientStockError struct {
	Items []InsufficientStockItem `json:"items"`
}

func (e *InsufficientStockError) Error() string {
	return "insufficient stock for one or more items"
}
