package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type AddressRepository interface {
	CreateAddress(ctx context.Context, address *types.Address) error
	UpdateAddress(ctx context.Context, userID string, product types.Address) error
	RemoveAddress(ctx context.Context, userID, addressID string) error
}

type addressRepository struct {
	db *sql.DB
}

func NewAddressRepository(db *sql.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) CreateAddress(ctx context.Context, address *types.Address) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if existing address exists
	var addressID string
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM addresses
		WHERE user_id = $1 AND
			addressee = $2 AND
			line1 = $3 AND
			line2 = $4 AND
			city = $5 AND
			state = $6 AND
			postal_code = $7 AND
			country = $8 AND
			is_deleted = FALSE
	`,
		address.UserID,
		address.Addressee,
		address.Line1,
		address.Line2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
	).Scan(&addressID)
	if err == sql.ErrNoRows {
		addressID = ""
	} else if err != nil {
		return err
	}

	// if true, return existing address
	if addressID != "" {
		address.ID = addressID
		return tx.Commit()
	}

	// if false, create new address
	if err := r.createAddress(ctx, tx, address); err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

func (r *addressRepository) UpdateAddress(ctx context.Context, userID string, address types.Address) error {
	query := `UPDATE addresses SET
		addressee = $1,
		line1 = $2,
		line2 = $3,
		city = $4,
		state = $5,
		postal_code = $6,
		country = $7,
		updated_at = NOW()
		WHERE user_id = $8 AND id = $9
	`
	res, err := r.db.ExecContext(ctx, query,
		address.Addressee,
		address.Line1,
		address.Line2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		userID,
		address.ID,
	)
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

func (r *addressRepository) createAddress(ctx context.Context, tx *sql.Tx, address *types.Address) error {
	query := `
		INSERT INTO addresses (
			id,
			user_id,
			addressee,
			line1,
			line2,
			city,
			state,
			postal_code,
			country
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, created_at
	`

	return tx.QueryRowContext(ctx, query,
		address.ID,
		address.UserID,
		address.Addressee,
		address.Line1,
		address.Line2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
	).Scan(&address.ID, &address.UserID, &address.CreatedAt)
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
