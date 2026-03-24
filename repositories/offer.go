package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type OfferRepository interface {
	CreateOffer(ctx context.Context, offer *types.Offer) error
	UpdateOffer(ctx context.Context, offer *types.Offer) error
	GetOfferByID(ctx context.Context, id string) (types.Offer, error)
	GetOffersByProductIDAndUser(ctx context.Context, productID, userID string) ([]types.Offer, error)
	GetOffers(ctx context.Context) ([]types.Offer, error)
}

type offerRepository struct {
	db *sql.DB
}

func NewOfferRepository(db *sql.DB) OfferRepository {
	return &offerRepository{db: db}
}

func (r *offerRepository) CreateOffer(ctx context.Context, offer *types.Offer) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Abort if user has pending offer
	var pendingExists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM offers
			WHERE user_id = $1
				AND product_id = $2
				AND status = 'pending'
		)
	`, offer.UserID, offer.Product.ID).Scan(&pendingExists)
	if err != nil {
		return err
	}
	if pendingExists {
		return types.ErrConstraintViolation
	}

	// Lock product row and check inventory atomically
	var inventory int
	err = tx.QueryRowContext(ctx, `SELECT inventory FROM products WHERE id = $1 AND (price = 0 OR pickup_only = true) FOR UPDATE`, offer.Product.ID).
		Scan(&inventory)
	if err == sql.ErrNoRows {
		return types.ErrNotFound
	}
	if err != nil {
		return err
	}
	if inventory < 1 {
		return types.ErrConstraintViolation
	}

	// decrement inventory if offer has been accepted
	if offer.Status == types.OfferAccepted {
		if _, err := tx.ExecContext(ctx, "UPDATE products SET inventory = inventory - 1 WHERE id = $1", offer.Product.ID); err != nil {
			return err
		}
	}

	// Insert record into offers
	query := `
		INSERT INTO offers (id, user_id, product_id, amount, status, pickup_notes)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := tx.ExecContext(ctx,
		query,
		offer.ID,
		offer.UserID,
		offer.Product.ID,
		offer.Amount,
		offer.Status,
		offer.PickupNotes,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *offerRepository) UpdateOffer(ctx context.Context, offer *types.Offer) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// decrement inventory when offer has been accepted
	if offer.Status == types.OfferAccepted {
		// get the product ID
		offer.Product = types.Product{}
		err := r.db.QueryRowContext(ctx, "SELECT product_id FROM offers WHERE id = $1", offer.ID).Scan(&offer.Product.ID)
		if err == sql.ErrNoRows {
			return types.ErrNotFound
		}

		// Lock product row and check inventory atomically
		var inventory int
		err = tx.QueryRowContext(ctx, `SELECT inventory FROM products WHERE id = $1 AND (price = 0 OR pickup_only = true) FOR UPDATE`, offer.Product.ID).
			Scan(&inventory)
		if err == sql.ErrNoRows {
			return types.ErrNotFound
		}
		if err != nil {
			return err
		}
		if inventory < 1 {
			return types.ErrConstraintViolation
		}

		if _, err := tx.ExecContext(ctx, "UPDATE products SET inventory = inventory - 1 WHERE id = $1", offer.Product.ID); err != nil {
			return err
		}
	}

	query := `
		UPDATE offers SET status = $2
		WHERE id = $1
		RETURNING product_id, user_id
	`
	err = tx.QueryRowContext(ctx, query, offer.ID, offer.Status).Scan(&offer.Product.ID, &offer.UserID)
	if err == sql.ErrNoRows {
		return types.ErrNotFound
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *offerRepository) GetOffersByProductIDAndUser(ctx context.Context, productID, userID string) ([]types.Offer, error) {
	offers := []types.Offer{}
	query := `
	SELECT 
		id,
		user_id,
		product_id,
		amount,
		status,
		pickup_notes,
		created_at,
		updated_at
	FROM offers
	WHERE product_id = $1 AND user_id = $2
	`
	rows, err := r.db.QueryContext(ctx, query, productID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pi types.Offer
		pi.Product = types.Product{}
		if err = rows.Scan(
			&pi.ID,
			&pi.UserID,
			&pi.Product.ID,
			&pi.Amount,
			&pi.Status,
			&pi.PickupNotes,
			&pi.CreatedAt,
			&pi.UpdatedAt,
		); err != nil {
			return nil, err
		}

		offers = append(offers, pi)
	}

	return offers, nil
}

func (r *offerRepository) GetOfferByID(ctx context.Context, id string) (types.Offer, error) {
	query := `
	SELECT
		id,
		user_id,
		product_id,
		amount,
		status,
		pickup_notes,
		created_at,
		updated_at
	FROM offers
	WHERE id = $1
	`
	var pi types.Offer
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pi.ID,
		&pi.UserID,
		&pi.Product.ID,
		&pi.Amount,
		&pi.Status,
		&pi.PickupNotes,
		&pi.CreatedAt,
		&pi.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return pi, types.ErrNotFound
	}
	return pi, err
}

func (r *offerRepository) GetOffers(ctx context.Context) ([]types.Offer, error) {
	offers := []types.Offer{}
	query := `
	SELECT
		id,
		user_id,
		product_id,
		amount,
		status,
		pickup_notes,
		created_at,
		updated_at
	FROM offers
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pi types.Offer
		pi.Product = types.Product{}
		if err = rows.Scan(
			&pi.ID,
			&pi.UserID,
			&pi.Product.ID,
			&pi.Amount,
			&pi.Status,
			&pi.PickupNotes,
			&pi.CreatedAt,
			&pi.UpdatedAt,
		); err != nil {
			return nil, err
		}

		offers = append(offers, pi)
	}

	return offers, nil
}
