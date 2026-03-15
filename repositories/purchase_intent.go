package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type PurchaseIntentRepository interface {
	CreatePurchaseIntent(ctx context.Context, purchaseIntent *types.PurchaseIntent) error
	UpdatePurchaseIntent(ctx context.Context, purchaseIntent *types.PurchaseIntent) error
	GetPurchaseIntentByID(ctx context.Context, id string) (types.PurchaseIntent, error)
	GetPurchaseIntentsByProductIDAndUser(ctx context.Context, productID, userID string) ([]types.PurchaseIntent, error)
	GetPurchaseIntents(ctx context.Context) ([]types.PurchaseIntent, error)
}

type purchaseIntentRepository struct {
	db *sql.DB
}

func NewPurchaseIntentRepository(db *sql.DB) PurchaseIntentRepository {
	return &purchaseIntentRepository{db: db}
}

func (r *purchaseIntentRepository) CreatePurchaseIntent(ctx context.Context, purchaseIntent *types.PurchaseIntent) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Abort if user has pending purchase intent
	var pendingExists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM purchase_intents
			WHERE user_id = $1
				AND product_id = $2
				AND status = 'pending'
		)
	`, purchaseIntent.UserID, purchaseIntent.Product.ID).Scan(&pendingExists)
	if err != nil {
		return err
	}
	if pendingExists {
		return types.ErrConstraintViolation
	}

	// Lock product row and check inventory atomically
	var inventory int
	err = tx.QueryRowContext(ctx, `SELECT inventory FROM products WHERE id = $1 AND (price = 0 OR pickup_only = true) FOR UPDATE`, purchaseIntent.Product.ID).
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

	// decrement inventory if purchase intent has been accepted
	if purchaseIntent.Status == types.PurchaseIntentAccepted {
		if _, err := tx.ExecContext(ctx, "UPDATE products SET inventory = inventory - 1 WHERE id = $1", purchaseIntent.Product.ID); err != nil {
			return err
		}
	}

	// Insert record into purchase_intents
	query := `
		INSERT INTO purchase_intents (id, user_id, product_id, offer_price, status, pickup_notes)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := tx.ExecContext(ctx,
		query,
		purchaseIntent.ID,
		purchaseIntent.UserID,
		purchaseIntent.Product.ID,
		purchaseIntent.OfferPrice,
		purchaseIntent.Status,
		purchaseIntent.PickupNotes,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *purchaseIntentRepository) UpdatePurchaseIntent(ctx context.Context, purchaseIntent *types.PurchaseIntent) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// decrement inventory when purchase intent has been accepted
	if purchaseIntent.Status == types.PurchaseIntentAccepted {
		// get the product ID
		purchaseIntent.Product = types.Product{}
		err := r.db.QueryRowContext(ctx, "SELECT product_id FROM purchase_intents WHERE id = $1", purchaseIntent.ID).Scan(&purchaseIntent.Product.ID)
		if err == sql.ErrNoRows {
			return types.ErrNotFound
		}

		// Lock product row and check inventory atomically
		var inventory int
		err = tx.QueryRowContext(ctx, `SELECT inventory FROM products WHERE id = $1 AND (price = 0 OR pickup_only = true) FOR UPDATE`, purchaseIntent.Product.ID).
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

		if _, err := tx.ExecContext(ctx, "UPDATE products SET inventory = inventory - 1 WHERE id = $1", purchaseIntent.Product.ID); err != nil {
			return err
		}
	}

	query := `
		UPDATE purchase_intents SET status = $2
		WHERE id = $1
		RETURNING product_id, user_id
	`
	err = tx.QueryRowContext(ctx, query, purchaseIntent.ID, purchaseIntent.Status).Scan(&purchaseIntent.Product.ID, &purchaseIntent.UserID)
	if err == sql.ErrNoRows {
		return types.ErrNotFound
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *purchaseIntentRepository) GetPurchaseIntentsByProductIDAndUser(ctx context.Context, productID, userID string) ([]types.PurchaseIntent, error) {
	purchaseIntents := []types.PurchaseIntent{}
	query := `
	SELECT 
		id,
		user_id,
		product_id,
		offer_price,
		status,
		pickup_notes,
		created_at,
		updated_at
	FROM purchase_intents
	WHERE product_id = $1 AND user_id = $2
	`
	rows, err := r.db.QueryContext(ctx, query, productID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pi types.PurchaseIntent
		pi.Product = types.Product{}
		if err = rows.Scan(
			&pi.ID,
			&pi.UserID,
			&pi.Product.ID,
			&pi.OfferPrice,
			&pi.Status,
			&pi.PickupNotes,
			&pi.CreatedAt,
			&pi.UpdatedAt,
		); err != nil {
			return nil, err
		}

		purchaseIntents = append(purchaseIntents, pi)
	}

	return purchaseIntents, nil
}

func (r *purchaseIntentRepository) GetPurchaseIntentByID(ctx context.Context, id string) (types.PurchaseIntent, error) {
	query := `
	SELECT
		id,
		user_id,
		product_id,
		offer_price,
		status,
		pickup_notes,
		created_at,
		updated_at
	FROM purchase_intents
	WHERE id = $1
	`
	var pi types.PurchaseIntent
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pi.ID,
		&pi.UserID,
		&pi.Product.ID,
		&pi.OfferPrice,
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

func (r *purchaseIntentRepository) GetPurchaseIntents(ctx context.Context) ([]types.PurchaseIntent, error) {
	purchaseIntents := []types.PurchaseIntent{}
	query := `
	SELECT
		id,
		user_id,
		product_id,
		offer_price,
		status,
		pickup_notes,
		created_at,
		updated_at
	FROM purchase_intents
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pi types.PurchaseIntent
		pi.Product = types.Product{}
		if err = rows.Scan(
			&pi.ID,
			&pi.UserID,
			&pi.Product.ID,
			&pi.OfferPrice,
			&pi.Status,
			&pi.PickupNotes,
			&pi.CreatedAt,
			&pi.UpdatedAt,
		); err != nil {
			return nil, err
		}

		purchaseIntents = append(purchaseIntents, pi)
	}

	return purchaseIntents, nil
}
