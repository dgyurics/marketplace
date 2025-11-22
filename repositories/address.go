package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type AddressRepository interface {
	CreateAddress(ctx context.Context, address *types.Address) error
	UpdateAddress(ctx context.Context, address *types.Address) error
	GetAddress(ctx context.Context, userID, addressID string) (types.Address, error)
	RemoveAddress(ctx context.Context, userID, addressID string) error
}

type addressRepository struct {
	db *sql.DB
}

func NewAddressRepository(db *sql.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) GetAddress(ctx context.Context, userID, addressID string) (types.Address, error) {
	var addr types.Address
	query := `
		SELECT id, user_id, name, line1, line2,
			city, state, postal_code, country, email, created_at, updated_at
		FROM addresses
		WHERE id = $1 AND user_id = $2
	`
	err := r.db.QueryRowContext(ctx, query, addressID, userID).Scan(
		&addr.ID, &addr.UserID, &addr.Name, &addr.Line1, &addr.Line2,
		&addr.City, &addr.State, &addr.PostalCode, &addr.Country, &addr.Email, &addr.CreatedAt, &addr.UpdatedAt)
	if err == sql.ErrNoRows {
		return addr, types.ErrNotFound
	}
	return addr, err
}

func (r *addressRepository) UpdateAddress(ctx context.Context, address *types.Address) error {
	query := `UPDATE addresses SET
		name = $1,
		line1 = $2,
		line2 = $3,
		city = $4,
		state = $5,
		postal_code = $6,
		country = $7,
		email = $8,
		updated_at = NOW()
		WHERE user_id = $9 AND id = $10
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		address.Name,
		address.Line1,
		address.Line2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.Email,
		address.UserID,
		address.ID,
	).Scan(&address.UpdatedAt)
	if err == sql.ErrNoRows {
		return types.ErrNotFound
	}
	return err
}

func (r *addressRepository) CreateAddress(ctx context.Context, address *types.Address) error {
	query := `
		INSERT INTO addresses (
			id,
			user_id,
			name,
			line1,
			line2,
			city,
			state,
			postal_code,
			country,
			email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at
	`
	return r.db.QueryRowContext(ctx, query,
		address.ID,
		address.UserID,
		address.Name,
		address.Line1,
		address.Line2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.Email,
	).Scan(&address.CreatedAt)
}

func (r *addressRepository) RemoveAddress(ctx context.Context, userID, addressID string) error {
	query := `
		DELETE FROM addresses
		WHERE id = $1 AND user_id = $2
	`
	res, err := r.db.ExecContext(ctx, query, addressID, userID)
	if err != nil {
		return err
	}
	// lib/pq always returns nil error for RowsAffected()
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return types.ErrNotFound
	}
	return nil
}
