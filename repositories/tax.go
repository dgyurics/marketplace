package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dgyurics/marketplace/types"
)

type TaxRepository interface {
	GetTaxRates(ctx context.Context, address types.Address, taxCode string) (int32, error)
}

type taxRepository struct {
	db *sql.DB
}

func NewTaxRepository(db *sql.DB) TaxRepository {
	return &taxRepository{db: db}
}

// GetTaxRates retrieves the tax rate for a given address and tax code.
// If taxCode is empty, it will return the default tax rate for the country and state (general goods and services).
// State is optional for countries that do not have state-level tax rates.
func (r *taxRepository) GetTaxRates(ctx context.Context, address types.Address, taxCode string) (int32, error) {
	query := `
		SELECT percentage
		FROM tax_rates
		WHERE country = $1
	`
	args := []interface{}{}
	argCount := 1

	args = append(args, strings.ToUpper(address.Country))
	argCount++

	if address.State != "" {
		query += fmt.Sprintf(" AND state = $%d", argCount)
		args = append(args, strings.ToUpper(address.State))
		argCount++
	}

	if taxCode == "" {
		query += " AND tax_code IS NULL"
	} else {
		query += fmt.Sprintf(" AND tax_code = $%d", argCount)
		args = append(args, taxCode)
		argCount++
	}

	var rate int32
	return rate, r.db.QueryRowContext(ctx, query, args...).Scan(&rate)
}
