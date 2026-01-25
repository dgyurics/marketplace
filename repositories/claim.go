package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type ClaimRepository interface {
	ClaimItem(ctx context.Context, claim *types.Claim) error
}

type claimRepository struct {
	db *sql.DB
}

func NewClaimRepository(db *sql.DB) ClaimRepository {
	return &claimRepository{db: db}
}

func (r *claimRepository) ClaimItem(ctx context.Context, claim *types.Claim) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock the product row and check inventory atomically
	var inventory int
	err = tx.QueryRowContext(ctx, `SELECT inventory FROM products WHERE id = $1 AND price = 0 FOR UPDATE`, claim.Product.ID).Scan(&inventory)
	if err == sql.ErrNoRows {
		return types.ErrNotFound
	}
	if err != nil {
		return err
	}
	if inventory < 1 {
		return types.ErrConstraintViolation
	}

	// Decrement product.inventory by 1
	if _, err := tx.ExecContext(ctx, "UPDATE products SET inventory = inventory - 1 WHERE id = $1", claim.Product.ID); err != nil {
		return err
	}

	// Insert record into claims
	if _, err := tx.ExecContext(ctx, "INSERT INTO claims (id, user_id, pickup_notes, product_id) VALUES ($1, $2, $3, $4)", claim.ID, claim.UserID, claim.PickupNotes, claim.Product.ID); err != nil {
		return err
	}

	return tx.Commit()
}
