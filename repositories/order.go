package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dgyurics/marketplace/types"
)

type OrderRepository interface {
	CancelPendingOrders(ctx context.Context, interval time.Duration) error
	CreateOrder(ctx context.Context, order *types.Order) error
	UpdateOrder(ctx context.Context, params types.OrderParams) (types.Order, error)
	MarkOrderAsPaid(ctx context.Context, orderID string) error
	/* GET order(s) */
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

func (r *orderRepository) CancelPendingOrders(ctx context.Context, interval time.Duration) error {
	intervalStr := fmt.Sprintf("%d seconds", int(interval.Seconds()))
	query := `
		UPDATE orders
		SET status = 'canceled', updated_at = CURRENT_TIMESTAMP
		WHERE status = 'pending' AND updated_at < NOW() - ($1)::INTERVAL
	`
	if _, err := r.db.ExecContext(ctx, query, intervalStr); err != nil {
		return err
	}
	return r.restockCanceledOrderItems(ctx)
}

func (r *orderRepository) restockCanceledOrderItems(ctx context.Context) error {
	query := `
		WITH deleted_items AS (
			DELETE FROM order_items oi
			USING orders o
			WHERE o.id = oi.order_id AND o.status = 'canceled'
			RETURNING oi.product_id, oi.quantity
		)
		UPDATE products p
		SET inventory = p.inventory + di.quantity
		FROM deleted_items di
		WHERE p.id = di.product_id;
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *types.Order) error {
	query := `
		INSERT INTO orders (id, user_id, address_id, status)
		VALUES ($1, $2, $3, 'created')
		RETURNING id, status, created_at
	`
	return r.db.QueryRowContext(ctx, query, order.ID, order.UserID, order.Address.ID).
		Scan(&order.ID, &order.Status, &order.CreatedAt)
}

func (r *orderRepository) UpdateOrder(ctx context.Context, params types.OrderParams) (ord types.Order, err error) {
	if params.ID == "" {
		return ord, errors.New("order ID is required")
	}
	if params.UserID == "" {
		return ord, errors.New("user ID is required")
	}

	query := `UPDATE orders SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}

	attrs := []slog.Attr{
		slog.String("order_id", params.ID),
	}

	if params.Status != nil {
		attrs = append(attrs, slog.String("status", string(*params.Status)))
		query += fmt.Sprintf(", status = $%d", len(args)+1)
		args = append(args, *params.Status)
	}

	if params.TaxAmount != nil {
		attrs = append(attrs, slog.Int64("tax_amount", *params.TaxAmount))
		query += fmt.Sprintf(", tax_amount = $%d", len(args)+1)
		args = append(args, *params.TaxAmount)
	}

	if params.ShippingAmount != nil {
		attrs = append(attrs, slog.Int64("shipping_amount", *params.ShippingAmount))
		query += fmt.Sprintf(", shipping_amount = $%d", len(args)+1)
		args = append(args, *params.ShippingAmount)
	}

	if params.TotalAmount != nil {
		attrs = append(attrs, slog.Int64("total_amount", *params.TotalAmount))
		query += fmt.Sprintf(", total_amount = $%d", len(args)+1)
		args = append(args, *params.TotalAmount)
	}

	if len(args) == 0 {
		return ord, fmt.Errorf("no fields to update")
	}

	query += fmt.Sprintf(" WHERE id = $%d", len(args)+1)
	args = append(args, params.ID)

	query += fmt.Sprintf(" AND user_id = $%d", len(args)+1)
	args = append(args, params.UserID)

	slog.LogAttrs(ctx, slog.LevelDebug, "Updating order", attrs...)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		slog.Error("Failed to update order", "error", err, "order_id", params.ID, "user_id", params.UserID)
		return ord, err
	}

	ord, err = r.GetOrderByIDAndUser(ctx, params.ID, params.UserID)
	if err != nil {
		slog.Error("Failed to retrieve updated order", "error", err, "order_id", params.ID, "user_id", params.UserID)
		return ord, err
	}

	// If the order was canceled, restock the items
	if ord.Status == types.OrderCanceled {
		return ord, r.restockCanceledOrderItems(ctx)
	}

	return ord, nil
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
			a.addressee,
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
			&order.Address.Addressee,
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
	if orderID == "" {
		return order, errors.New("order ID is required")
	}
	if userID == "" {
		return order, errors.New("user ID is required")
	}
	query := `
		SELECT
			o.id,
			o.user_id,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.address_id,
			a.addressee,
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
		&order.Address.Addressee,
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
			a.addressee,
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
		&order.Address.Addressee,
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

func (r *orderRepository) MarkOrderAsPaid(ctx context.Context, orderID string) error {
	query := `
		UPDATE orders
		SET
			status = 'paid',
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, orderID)
	return err
}
