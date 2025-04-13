package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dgyurics/marketplace/types"
)

type AddressRepository interface {
	CreateAddress(ctx context.Context, address *types.Address) error
	UpdateAddress(ctx context.Context, address *types.Address) error
	GetAddresses(ctx context.Context, userID string) ([]types.Address, error)
	RemoveAddress(ctx context.Context, userID, addressID string) error
}

type addressRepository struct {
	db *sql.DB
}

func (r *addressRepository) UpdateAddress(ctx context.Context, address *types.Address) error {
	query := `
		UPDATE addresses
		SET
			addressee = $1,
			address_line1 = $2,
			address_line2 = $3,
			city = $4,
			state_code = $5,
			postal_code = $6,
			phone = $7,
			updated_at = NOW()
		WHERE id = $8 AND user_id = $9
		RETURNING updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		address.Addressee,
		address.AddressLine1,
		address.AddressLine2,
		address.City,
		address.StateCode,
		address.PostalCode,
		address.Phone,
		address.ID,
		address.UserID,
	).Scan(&address.UpdatedAt)
}

func NewAddressRepository(db *sql.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) CreateAddress(ctx context.Context, address *types.Address) error {
	query := `
		INSERT INTO addresses (
			user_id,
			addressee,
			address_line1,
			address_line2,
			city,
			state_code,
			postal_code,
			country_code,
			phone
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, created_at, updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		address.UserID,
		address.Addressee,
		address.AddressLine1,
		address.AddressLine2,
		address.City,
		address.StateCode,
		address.PostalCode,
		address.CountryCode,
		address.Phone,
	).Scan(&address.ID, &address.UserID, &address.CreatedAt, &address.UpdatedAt)
}

func (r *addressRepository) GetAddresses(ctx context.Context, userID string) ([]types.Address, error) {
	query := `
		SELECT
			id,
			user_id,
			addressee,
			address_line1,
			address_line2,
			city,
			state_code,
			postal_code,
			country_code,
			phone,
			is_deleted,
			created_at,
			updated_at
		FROM addresses
		WHERE user_id = $1 AND is_deleted = FALSE
	`

	addresses := []types.Address{}
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address types.Address
		if err := rows.Scan(
			&address.ID,
			&address.UserID,
			&address.Addressee,
			&address.AddressLine1,
			&address.AddressLine2,
			&address.City,
			&address.StateCode,
			&address.PostalCode,
			&address.CountryCode,
			&address.Phone,
			&address.IsDeleted,
			&address.CreatedAt,
			&address.UpdatedAt,
		); err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

func (r *addressRepository) RemoveAddress(ctx context.Context, userID, addressID string) error {
	query := `
		DELETE FROM addresses
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, addressID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("address with ID %s not found", addressID)
	}

	return nil
}
