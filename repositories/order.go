package repositories

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"slices"

	"github.com/dgyurics/marketplace/types"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *types.Order) error
	UpdateOrder(ctx context.Context, order *types.Order) error
	GetOrderByIDAndUser(ctx context.Context, orderID, userID string) (types.Order, error)
	GetOrderByID(ctx context.Context, orderID string) (types.Order, error)
	GetOrderByIDPublic(ctx context.Context, orderID string) (types.Order, error)
	GetOrders(ctx context.Context, page, limit int) ([]types.Order, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *types.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := cancelPendingOrders(ctx, tx, order.UserID); err != nil {
		return err
	}

	// Reserve inventory (decrement stock, collect shortages)
	shortages, err := reserveInventory(ctx, tx, order.Items)
	if err != nil {
		return err
	}

	if len(shortages) > 0 {
		// Release/Restore inventory which was reserved
		if err := tx.Rollback(); err != nil {
			return err
		}
		// Update/Sync cart to reflect what is actually available
		if err := r.syncCartToInventory(ctx, order.UserID, shortages); err != nil {
			return err
		}
		return &types.InsufficientStockError{Items: shortages}
	}

	// Insert order, skipping if this idempotency key was already used
	inserted, err := insertOrder(ctx, tx, order)
	if err != nil {
		return err
	}
	if !inserted {
		// Duplicate request: return the original order's ID
		if order.IdempotencyKey == nil {
			return fmt.Errorf("order insert affected no rows without idempotency key")
		}
		return tx.QueryRowContext(ctx,
			`SELECT id FROM orders WHERE idempotency_key = $1`,
			*order.IdempotencyKey).Scan(&order.ID)
	}

	// Insert order items
	if err := insertOrderItems(ctx, tx, order.ID, order.Items); err != nil {
		return err
	}
	return tx.Commit()
}

// cancelPendingOrders cancels the user's outstanding draft, if any, and
// returns its reserved inventory to stock.
func cancelPendingOrders(ctx context.Context, tx *sql.Tx, userID string) error {
	_, err := tx.ExecContext(ctx, `
		WITH canceled AS (
			UPDATE orders SET status = 'canceled', updated_at = NOW()
			WHERE
				user_id = $1 AND status = 'draft'
			RETURNING id
		), restored AS (
			SELECT product_id, SUM(quantity) AS quantity
			FROM order_items
			WHERE order_id IN (SELECT id FROM canceled)
			GROUP BY product_id
		)
		UPDATE products
		SET inventory = inventory + restored.quantity, updated_at = NOW()
		FROM restored
		WHERE products.id = restored.product_id`,
		userID)
	return err
}

// reserveInventory decrements stock for each item. Items without sufficient
// stock are left untouched and returned as shortages, along with the quantity
// currently available.
func reserveInventory(ctx context.Context, tx *sql.Tx, items []types.OrderItem) ([]types.InsufficientStockItem, error) {
	var shortages []types.InsufficientStockItem

	// Lock product rows in a consistent order so concurrent orders containing
	// the same products cannot deadlock against each other.
	sorted := slices.Clone(items)
	slices.SortFunc(sorted, func(a, b types.OrderItem) int {
		return cmp.Compare(a.Product.ID, b.Product.ID)
	})

	for _, item := range sorted {
		res, err := tx.ExecContext(ctx, `
			UPDATE products
			SET inventory = inventory - $1, updated_at = NOW()
			WHERE id = $2 AND inventory >= $1`,
			item.Quantity, item.Product.ID)
		if err != nil {
			return nil, err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return nil, err
		}
		if rows > 0 {
			continue
		}

		var inventory int
		if err := tx.QueryRowContext(ctx,
			`SELECT inventory FROM products WHERE id = $1`,
			item.Product.ID).Scan(&inventory); err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("product %s not found", item.Product.ID)
			}
			return nil, err
		}
		shortages = append(shortages, types.InsufficientStockItem{
			Product:   item.Product,
			Quantity:  item.Quantity,
			Inventory: inventory,
		})
	}
	return shortages, nil
}

// insertOrder writes the order row. It reports false without error when the
// idempotency key has already been used, meaning this is a duplicate request.
func insertOrder(ctx context.Context, tx *sql.Tx, order *types.Order) (bool, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO orders (
			id,
			user_id,
			address_id,
			amount,
			tax_amount,
			shipping_amount,
			total_amount,
			idempotency_key,
			status,
			payment_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL
		DO NOTHING`,
		order.ID,
		order.UserID,
		order.Address.ID,
		order.Amount,
		order.TaxAmount,
		order.ShippingAmount,
		order.TotalAmount,
		order.IdempotencyKey,
		order.Status,
		order.PaymentStatus,
	)
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// insertOrderItems writes the line items belonging to an order.
func insertOrderItems(ctx context.Context, tx *sql.Tx, orderID string, items []types.OrderItem) error {
	for _, item := range items {
		// FIXME convert to multi-row insert
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, unit_price)
			VALUES ($1, $2, $3, $4)`,
			orderID, item.Product.ID, item.Quantity, item.UnitPrice); err != nil {
			return err
		}
	}
	return nil
}

// syncCartToInventory trims the user's cart to the quantities actually in
// stock, removing items which are unavailable
func (r *orderRepository) syncCartToInventory(ctx context.Context, userID string, shortages []types.InsufficientStockItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range shortages {
		if item.Inventory <= 0 {
			_, err = tx.ExecContext(ctx, `
				DELETE FROM cart_items
				WHERE user_id = $1 AND product_id = $2`,
				userID, item.Product.ID)
		} else {
			_, err = tx.ExecContext(ctx, `
				UPDATE cart_items SET quantity = $1
				WHERE user_id = $2 AND product_id = $3`,
				item.Inventory, userID, item.Product.ID)
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetOrders retrieves all orders in descending order
func (r *orderRepository) GetOrders(ctx context.Context, page, limit int) ([]types.Order, error) {
	query := `
		SELECT
			o.id,
			o.user_id,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.payment_method,
			o.payment_status,
			a.id AS address_id,
			a.name,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			a.country,
			o.created_at,
			o.updated_at
		FROM orders o
		JOIN addresses a ON o.address_id = a.id
		ORDER BY o.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []types.Order{}
	for rows.Next() {
		order := types.Order{
			Address: types.Address{},
		}

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Amount,
			&order.TaxAmount,
			&order.TotalAmount,
			&order.Status,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.Address.ID,
			&order.Address.Name,
			&order.Address.Line1,
			&order.Address.Line2,
			&order.Address.City,
			&order.Address.State,
			&order.Address.PostalCode,
			&order.Address.Country,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, order)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// populateOrderItems populates the order items for a list of orders
func (r *orderRepository) populateOrderItems(ctx context.Context, orderID string) ([]types.OrderItem, error) {
	query := `
		SELECT
			product_id,
			name,
			summary,
			thumbnail,
			alt_text,
			quantity,
			unit_price
		FROM v_order_items
		WHERE order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process query results
	items := []types.OrderItem{}
	for rows.Next() {
		item := types.OrderItem{}
		if err := rows.Scan(
			&item.Product.ID,
			&item.Product.Name,
			&item.Product.Summary,
			&item.Thumbnail,
			&item.AltText,
			&item.Quantity,
			&item.UnitPrice,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *orderRepository) GetOrderByIDAndUser(ctx context.Context, orderID, userID string) (types.Order, error) {
	var order types.Order
	query := `
		SELECT
			o.id,
			o.user_id,
			o.amount,
			o.tax_amount,
			o.shipping_amount,
			o.total_amount,
			o.status,
			o.payment_method,
			o.payment_status,
			o.address_id,
			a.name,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			a.country,
			o.created_at,
			o.updated_at
		FROM orders o
		LEFT JOIN addresses a ON o.address_id = a.id
		WHERE
			o.id = $1 AND
			o.user_id = $2
	`
	order.Address = types.Address{}
	err := r.db.QueryRowContext(ctx, query, orderID, userID).Scan(
		&order.ID,
		&order.UserID,
		&order.Amount,
		&order.TaxAmount,
		&order.ShippingAmount,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.Address.ID,
		&order.Address.Name,
		&order.Address.Line1,
		&order.Address.Line2,
		&order.Address.City,
		&order.Address.State,
		&order.Address.PostalCode,
		&order.Address.Country,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return order, types.ErrNotFound
	}
	if err != nil {
		return order, err
	}

	// Populate order items
	if order.Items, err = r.populateOrderItems(ctx, order.ID); err != nil {
		return order, fmt.Errorf("failed to populate order items: %w", err)
	}

	return order, nil
}

func (r *orderRepository) GetOrderByIDPublic(ctx context.Context, orderID string) (types.Order, error) {
	var order types.Order
	query := `
		SELECT
			o.id,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.created_at,
			o.updated_at
		FROM orders o
		WHERE o.id = $1
	`
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return order, types.ErrNotFound
	}
	if err != nil {
		return order, err
	}

	// Populate order items for this order
	if order.Items, err = r.populateOrderItems(ctx, order.ID); err != nil {
		return order, fmt.Errorf("failed to populate order items: %w", err)
	}

	return order, nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, orderID string) (types.Order, error) {
	var order types.Order
	order.Address = types.Address{}
	query := `
		SELECT
			o.id,
			o.user_id,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.payment_method,
			o.payment_status,
			o.address_id,
			a.name,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			a.country,
			o.created_at,
			o.updated_at
		FROM orders o
		LEFT JOIN addresses a ON o.address_id = a.id
		WHERE o.id = $1
	`
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.Address.ID,
		&order.Address.Name,
		&order.Address.Line1,
		&order.Address.Line2,
		&order.Address.City,
		&order.Address.State,
		&order.Address.PostalCode,
		&order.Address.Country,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return order, types.ErrNotFound
	}
	if err != nil {
		return order, err
	}

	// Populate order items for this order
	if order.Items, err = r.populateOrderItems(ctx, order.ID); err != nil {
		return order, fmt.Errorf("failed to populate order items: %w", err)
	}

	return order, nil
}

func (r *orderRepository) UpdateOrder(ctx context.Context, order *types.Order) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var previousStatus types.OrderStatus
	var previousMethod, paymentReference sql.NullString
	var previousPaymentStatus types.PaymentStatus
	err = tx.QueryRowContext(ctx, `
		SELECT status, payment_method, payment_reference, payment_status
		FROM orders WHERE id = $1 FOR UPDATE`, order.ID).
		Scan(&previousStatus, &previousMethod, &paymentReference, &previousPaymentStatus)
	if err == sql.ErrNoRows {
		return types.ErrNotFound
	}
	if err != nil {
		return err
	}

	paymentPending := previousStatus == types.OrderPending && previousMethod.String == string(types.PaymentMethodStripe)
	if order.PaymentReference != nil &&
		(!paymentPending || *order.PaymentReference == "" || (paymentReference.Valid && paymentReference.String != *order.PaymentReference)) {
		return types.ErrConstraintViolation
	}
	paymentConfirmed := previousPaymentStatus == types.PaymentStatusUnpaid &&
		order.Status != nil && *order.Status == types.OrderPaid && order.PaymentReference != nil
	if paymentPending && order.Status != nil && *order.Status != types.OrderPending && !paymentConfirmed {
		return types.ErrConstraintViolation
	}
	if order.PaymentMethod != nil && *order.PaymentMethod == types.PaymentMethodStripe &&
		order.Status != nil && *order.Status == types.OrderPending && previousStatus != types.OrderDraft {
		return types.ErrConstraintViolation
	}
	if order.PaymentMethod != nil && string(*order.PaymentMethod) != previousMethod.String &&
		previousStatus != types.OrderDraft {
		return types.ErrConstraintViolation
	}

	query := `
		UPDATE orders SET 
			status = COALESCE($1, status),
			payment_method = COALESCE($2, payment_method),
			payment_status = COALESCE($3, payment_status),
			amount = COALESCE($4, amount),
			tax_amount = COALESCE($5, tax_amount),
			shipping_amount = COALESCE($6, shipping_amount),
			total_amount = COALESCE($7, total_amount),
			payment_reference = COALESCE($8, payment_reference),
			updated_at = NOW()
		WHERE id = $9
		RETURNING user_id
	`
	err = tx.QueryRowContext(
		ctx,
		query,
		order.Status,
		order.PaymentMethod,
		order.PaymentStatus,
		order.Amount,
		order.TaxAmount,
		order.ShippingAmount,
		order.TotalAmount,
		order.PaymentReference,
		order.ID,
	).Scan(&order.UserID)
	if err == sql.ErrNoRows {
		return types.ErrNotFound
	}
	if err != nil {
		return err
	}

	if order.Status == nil || *order.Status == previousStatus {
		return tx.Commit()
	}

	if *order.Status == types.OrderRequested || *order.Status == types.OrderRefunded || *order.Status == types.OrderCanceled {
		if err := reclaimOrderItems(ctx, tx, order.ID, previousStatus); err != nil {
			return err
		}
		return tx.Commit()
	}

	if previousStatus != types.OrderRequested {
		return tx.Commit()
	}

	items, err := r.populateOrderItems(ctx, order.ID)
	if err != nil {
		return err
	}
	shortages, err := reserveInventory(ctx, tx, items)
	if err != nil {
		return err
	}
	if len(shortages) > 0 {
		return &types.InsufficientStockError{Items: shortages}
	}

	return tx.Commit()
}

// reclaimOrderItems releases reserved stock while preserving the order's items.
func reclaimOrderItems(ctx context.Context, tx *sql.Tx, orderID string, previousStatus types.OrderStatus) error {
	switch previousStatus {
	case types.OrderRequested, types.OrderCanceled, types.OrderRefunded:
		return nil
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE products
		SET inventory = inventory + oi.quantity, updated_at = NOW()
		FROM order_items oi
		WHERE products.id = oi.product_id AND oi.order_id = $1`, orderID)
	return err
}
