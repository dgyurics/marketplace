package repositories

import (
	"context"
	"database/sql"
	"fmt"

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
	// Check for existing order with same idempotency key
	if order.IdempotencyKey != nil {
		var existingID string
		err := r.db.QueryRowContext(ctx,
			`SELECT id FROM orders WHERE idempotency_key = $1`,
			*order.IdempotencyKey).Scan(&existingID)
		if err == nil {
			order.ID = existingID
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Cancel any existing pending order and restore its inventory
	_, err = tx.ExecContext(ctx, `
		WITH canceled AS (
			UPDATE orders SET status = 'canceled', updated_at = NOW()
			WHERE user_id = $1 AND status = 'pending'
			RETURNING id
		), restored AS (
			DELETE FROM order_items
			WHERE order_id IN (SELECT id FROM canceled)
			RETURNING product_id, quantity
		)
		UPDATE products
		SET inventory = inventory + restored.quantity
		FROM restored
		WHERE products.id = restored.product_id`,
		order.UserID)
	if err != nil {
		return err
	}

	// Reserve inventory (decrement stock, fail if insufficient)
	var insufStockErr types.InsufficientStockError
	for _, item := range order.Items {
		res, err := tx.ExecContext(ctx, `
			UPDATE products
			SET inventory = inventory - $1
			WHERE id = $2 AND inventory >= $1`,
			item.Quantity, item.Product.ID)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			var inventory int
			_ = tx.QueryRowContext(ctx,
				`SELECT inventory FROM products WHERE id = $1`,
				item.Product.ID).Scan(&inventory)
			insufStockErr.Items = append(insufStockErr.Items, types.InsufficientStockItem{
				Product:   item.Product,
				Quantity:  item.Quantity,
				Inventory: inventory,
			})
		}
	}

	if len(insufStockErr.Items) > 0 {
		// Adjust cart to reflect available inventory
		for _, item := range insufStockErr.Items {
			if item.Inventory <= 0 {
				tx.ExecContext(ctx, `
                    DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`,
					order.UserID, item.Product.ID)
			} else {
				tx.ExecContext(ctx, `
                    UPDATE cart_items SET quantity = $1
                    WHERE user_id = $2 AND product_id = $3`,
					item.Inventory, order.UserID, item.Product.ID)
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return &insufStockErr
	}

	// Insert order
	query := `
		INSERT INTO orders (id, user_id, address_id, amount, tax_amount, shipping_amount, total_amount, status, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8)`
	if _, err := tx.ExecContext(ctx, query, order.ID, order.UserID, order.Address.ID, order.Amount,
		order.TaxAmount, order.ShippingAmount, order.TotalAmount, order.IdempotencyKey); err != nil {
		return err
	}

	// Insert order items
	for _, item := range order.Items {
		itemQuery := `
			INSERT INTO order_items (order_id, product_id, quantity, unit_price)
			VALUES ($1, $2, $3, $4)`
		if _, err := tx.ExecContext(ctx, itemQuery, order.ID, item.Product.ID, item.Quantity, item.UnitPrice); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *orderRepository) populateInsufficientStockError(ctx context.Context, item *types.InsufficientStockItem) error {
	return nil
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
			a.id AS address_id,
			a.name,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			a.country,
			a.email,
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
			&order.Address.ID,
			&order.Address.Name,
			&order.Address.Line1,
			&order.Address.Line2,
			&order.Address.City,
			&order.Address.State,
			&order.Address.PostalCode,
			&order.Address.Country,
			&order.Address.Email,
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
			o.total_amount,
			o.status,
			o.address_id,
			a.name,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			a.country,
			a.email,
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
		&order.TotalAmount,
		&order.Status,
		&order.Address.ID,
		&order.Address.Name,
		&order.Address.Line1,
		&order.Address.Line2,
		&order.Address.City,
		&order.Address.State,
		&order.Address.PostalCode,
		&order.Address.Country,
		&order.Address.Email,
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
			o.address_id,
			a.name,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			a.country,
			a.email,
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
		&order.Address.ID,
		&order.Address.Name,
		&order.Address.Line1,
		&order.Address.Line2,
		&order.Address.City,
		&order.Address.State,
		&order.Address.PostalCode,
		&order.Address.Country,
		&order.Address.Email,
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

	query := `
		UPDATE orders SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING user_id
	`
	if err := tx.QueryRowContext(ctx, query, order.Status, order.ID).Scan(&order.UserID); err != nil {
		return err
	}

	// restock inventory
	if order.Status == types.OrderRefunded ||
		order.Status == types.OrderCanceled {
		query = `
			WITH deleted_items AS (
				DELETE FROM order_items oi
				WHERE oi.order_id = $1
				RETURNING oi.product_id, oi.quantity
			)
			UPDATE products
			SET inventory = inventory + di.quantity
			FROM deleted_items di
			WHERE products.id = di.product_id
		`
		if _, err := tx.ExecContext(ctx, query, order.ID); err != nil {
			return err
		}
	}

	// clear cart
	if order.Status == types.OrderPaid {
		query = `
			WITH ordered AS (
				SELECT product_id, quantity
				FROM order_items
				WHERE order_id = $2
			)
			UPDATE cart_items ci
			SET quantity = ci.quantity - o.quantity
			FROM ordered o
			WHERE ci.user_id = $1
			AND ci.product_id = o.product_id
		`
		if _, err := tx.ExecContext(ctx, query, order.UserID, order.ID); err != nil {
			return err
		}

		// remove cart items with zero or negative quantity
		query = `DELETE FROM cart_items WHERE user_id = $1 AND quantity <= 0`
		if _, err := tx.ExecContext(ctx, query, order.UserID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
