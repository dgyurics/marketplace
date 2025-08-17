package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
)

type OrderRepository interface {
	CancelPendingOrders(ctx context.Context, interval time.Duration) ([]string, error) // TODO refactor, returning just stripe IDs is hacky
	CreateOrder(ctx context.Context, order *types.Order) error
	GetOrder(ctx context.Context, orderID, userID string) (types.Order, error)
	GetPendingOrder(ctx context.Context, userID string) (types.Order, error)
	GetOrders(ctx context.Context, userID string, page, limit int) ([]types.Order, error)
	UpdateOrder(ctx context.Context, params types.OrderParams) (types.Order, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CancelPendingOrders(ctx context.Context, interval time.Duration) ([]string, error) {
	intervalStr := fmt.Sprintf("%d seconds", int(interval.Seconds()))
	query := `
		UPDATE orders
		SET status = 'canceled', updated_at = CURRENT_TIMESTAMP
		WHERE status = 'pending' AND updated_at < NOW() - ($1)::INTERVAL
		RETURNING stripe_payment_intent->>'id' AS payment_intent_id
	`
	rows, err := r.db.QueryContext(ctx, query, intervalStr)
	if err != nil {
		return nil, err
	}
	var stripeIDs []string
	for rows.Next() {
		var id sql.NullString
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if id.Valid {
			stripeIDs = append(stripeIDs, id.String)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stripeIDs, r.restockCanceledOrderItems(ctx)
}

func (r *orderRepository) restockCanceledOrderItems(ctx context.Context) error {
	query := `
		WITH deleted_items AS (
			DELETE FROM order_items oi
			USING orders o
			WHERE o.id = oi.order_id AND o.status = 'canceled'
			RETURNING oi.product_id, oi.quantity
		)
		UPDATE inventory i
		SET quantity = i.quantity + di.quantity
		FROM deleted_items di
		WHERE i.product_id = di.product_id;
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// CreateOrder creates a new order from the user's cart items
func (r *orderRepository) CreateOrder(ctx context.Context, order *types.Order) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}
	if order.UserID == "" {
		return errors.New("user ID is required")
	}
	if order.Currency == "" {
		return errors.New("currency is required")
	}
	if order.ID == "" {
		return errors.New("order ID is required")
	}

	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Retrieve cart items
	var cartItems []types.CartItem
	query := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.tax_code,
			p.images,
			p.summary,
			ci.quantity,
			ci.unit_price
		FROM cart_items ci
		JOIN v_products p ON ci.product_id = p.id
		WHERE ci.user_id = $1
	`
	rows, err := tx.QueryContext(ctx, query, order.UserID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item types.CartItem
		item.Product = types.Product{}
		var imagesJSON []byte

		if err = rows.Scan(
			&item.Product.ID,
			&item.Product.Name,
			&item.Product.Price,
			&item.Product.TaxCode,
			&imagesJSON,
			&item.Product.Summary,
			&item.Quantity,
			&item.UnitPrice,
		); err != nil {
			return err
		}

		// Convert JSON array to Go struct
		// FIXME seems counterintuitive to convert images to JSON in view/database
		// and then convert back to Go struct/array
		if err := json.Unmarshal(imagesJSON, &item.Product.Images); err != nil {
			return err
		}

		cartItems = append(cartItems, item)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	if len(cartItems) == 0 {
		slog.Debug("CreateOrder: no items in cart", "user_id", order.UserID)
		return types.ErrNotFound
	}

	// Calculate cart total, excluding tax + shipping.
	// Tax and shipping will be calculated later.
	amount := calculateOrderAmount(cartItems)

	// Reduce inventory
	if err = reduceInventory(ctx, tx, cartItems); err != nil {
		return err
	}

	// create a new order with pending status
	query = `
		INSERT INTO orders (id, user_id, currency, amount, total_amount) VALUES ($1, $2, $3, $4, $4)
		RETURNING id, user_id, currency, amount, total_amount, status, created_at
	`
	if err = tx.QueryRowContext(
		ctx, query,
		order.ID,
		order.UserID,
		order.Currency,
		amount,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.Currency,
		&order.Amount,
		&order.TotalAmount,
		&order.Status,
		&order.CreatedAt,
	); err != nil {
		return err
	}

	// Populate order_items table
	query = `
		INSERT INTO order_items (order_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
	`
	order.Items = make([]types.OrderItem, 0, len(cartItems))
	for _, item := range cartItems {
		if _, err = tx.ExecContext(ctx, query,
			order.ID,
			item.Product.ID,
			item.Quantity,
			item.UnitPrice,
		); err != nil {
			return err
		}
		order.Items = append(order.Items, types.OrderItem{
			Product:   item.Product,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	return tx.Commit()
}

func reduceInventory(ctx context.Context, tx *sql.Tx, items []types.CartItem) error {
	for _, item := range items {
		result, err := tx.ExecContext(ctx, `
			UPDATE inventory
			SET quantity = quantity - $1
			WHERE quantity >= $1 AND product_id = $2
		`, item.Quantity, item.Product.ID)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return fmt.Errorf("%s is out of stock or insufficient quantity remains", item.Product.ID)
		}
	}
	return nil
}

func calculateOrderAmount(items []types.CartItem) int64 {
	var total int64
	for _, item := range items {
		total += item.UnitPrice * int64(item.Quantity)
	}
	return total
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
	argCount := 1

	attrs := []slog.Attr{
		slog.String("order_id", params.ID),
	}

	if params.Status != nil {
		attrs = append(attrs, slog.String("status", string(*params.Status)))
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, *params.Status)
		argCount++
	}

	if params.AddressID != nil {
		attrs = append(attrs, slog.String("address_id", *params.AddressID))
		query += fmt.Sprintf(", address_id = $%d", argCount)
		args = append(args, *params.AddressID)
		argCount++
	}

	if params.TaxAmount != nil {
		attrs = append(attrs, slog.Int64("tax_amount", *params.TaxAmount))
		query += fmt.Sprintf(", tax_amount = $%d", argCount)
		args = append(args, *params.TaxAmount)
		argCount++
	}

	if params.ShippingAmount != nil {
		attrs = append(attrs, slog.Int64("shipping_amount", *params.ShippingAmount))
		query += fmt.Sprintf(", shipping_amount = $%d", argCount)
		args = append(args, *params.ShippingAmount)
		argCount++
	}

	if params.TotalAmount != nil {
		attrs = append(attrs, slog.Int64("total_amount", *params.TotalAmount))
		query += fmt.Sprintf(", total_amount = $%d", argCount)
		args = append(args, *params.TotalAmount)
		argCount++
	}

	if params.Email != nil {
		attrs = append(attrs, slog.String("email", *params.Email))
		query += fmt.Sprintf(", email = $%d", argCount)
		args = append(args, *params.Email)
		argCount++
	}

	if params.StripePaymentIntent != nil {
		intentJSON, err := json.Marshal(*params.StripePaymentIntent)
		if err != nil {
			return ord, fmt.Errorf("failed to encode stripe payment intent: %w", err)
		}

		attrs = append(attrs, slog.String("stripe_payment_intent", string(intentJSON)))
		query += fmt.Sprintf(", stripe_payment_intent = $%d", argCount)
		args = append(args, intentJSON)
		argCount++
	}

	if len(args) == 0 {
		return ord, fmt.Errorf("no fields to update")
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, params.ID)
	argCount++

	query += fmt.Sprintf(" AND user_id = $%d", argCount)
	args = append(args, params.UserID)
	argCount++

	slog.LogAttrs(ctx, slog.LevelDebug, "Updating order", attrs...)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		slog.Error("Failed to update order", "error", err, "order_id", params.ID, "user_id", params.UserID)
		return ord, err
	}

	ord, err = r.GetOrder(ctx, params.ID, params.UserID)
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

// GetOrders retrieves all orders for a user
func (r *orderRepository) GetOrders(ctx context.Context, userID string, page, limit int) ([]types.Order, error) {
	query := `
		SELECT
			o.id,
			o.user_id,
			o.email,
			o.currency,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.stripe_payment_intent,
			a.id AS address_id,
			a.addressee,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			o.created_at,
			o.updated_at
		FROM orders o
		JOIN addresses a ON o.address_id = a.id
		WHERE o.user_id = $1 AND o.status != 'pending'
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []types.Order{}
	for rows.Next() {
		var rawIntent []byte
		order := types.Order{
			Address: &types.Address{},
		}

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Email,
			&order.Currency,
			&order.Amount,
			&order.TaxAmount,
			&order.TotalAmount,
			&order.Status,
			&rawIntent,
			&order.Address.ID,
			&order.Address.Addressee,
			&order.Address.Line1,
			&order.Address.Line2,
			&order.Address.City,
			&order.Address.State,
			&order.Address.PostalCode,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(rawIntent) > 0 {
			var spi stripe.PaymentIntent
			if err := json.Unmarshal(rawIntent, &spi); err != nil {
				return nil, fmt.Errorf("failed to unmarshal Stripe payment intent: %w", err)
			}
			order.StripePaymentIntent = &spi
		}

		result = append(result, order)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Populate order items for each order
	for idx, order := range result {
		// TODO wrap in goroutine
		result[idx].Items, err = r.populateOrderItems(ctx, order.ID)
		if err != nil {
			slog.Error("Failed to populate order items", "order_id", order.ID, "error", err)
		}
	}

	return result, nil
}

// populateOrderItems populates the order items for a list of orders
func (r *orderRepository) populateOrderItems(ctx context.Context, orderID string) ([]types.OrderItem, error) {
	if orderID == "" {
		return nil, errors.New("order ID is required")
	}

	// Query to fetch order items
	// missing alt text, product name,
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

	// Query to fetch order items
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process query results
	var items []types.OrderItem
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

func (r *orderRepository) GetOrder(ctx context.Context, orderID, userID string) (types.Order, error) {
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
			COALESCE(o.email, '') AS email,
			o.currency,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.stripe_payment_intent,
			o.address_id,
			a.addressee,
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

	// Execute the query
	var rawIntent []byte
	var address struct {
		ID         sql.NullString
		Addressee  sql.NullString
		Line1      sql.NullString
		Line2      sql.NullString
		City       sql.NullString
		State      sql.NullString
		PostalCode sql.NullString
		Country    sql.NullString
	}
	err := r.db.QueryRowContext(ctx, query, orderID, userID).Scan(
		&order.ID,
		&order.UserID,
		&order.Email,
		&order.Currency,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&rawIntent,
		&address.ID,
		&address.Addressee,
		&address.Line1,
		&address.Line2,
		&address.City,
		&address.State,
		&address.PostalCode,
		&address.Country,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return order, types.ErrNotFound
	}
	if err != nil {
		return order, err
	}

	// Unmarshal the Stripe payment intent if it exists
	if len(rawIntent) > 0 {
		var payInt stripe.PaymentIntent
		err = json.Unmarshal(rawIntent, &payInt)
		if err != nil {
			return order, fmt.Errorf("failed to unmarshal Stripe payment intent: %w", err)
		}
		order.StripePaymentIntent = &payInt
	}

	// Populate address if it exists
	if address.ID.Valid {
		order.Address = &types.Address{
			ID:         address.ID.String,
			Addressee:  &address.Addressee.String,
			Line1:      address.Line1.String,
			Line2:      &address.Line2.String,
			City:       address.City.String,
			State:      address.State.String,
			PostalCode: address.PostalCode.String,
			Country:    address.Country.String,
		}
	}

	// Populate order items for this order
	if order.Items, err = r.populateOrderItems(ctx, order.ID); err != nil {
		return order, fmt.Errorf("failed to populate order items: %w", err)
	}

	return order, nil
}

func (r *orderRepository) GetPendingOrder(ctx context.Context, userID string) (types.Order, error) {
	var order types.Order
	if userID == "" {
		return order, errors.New("user ID is required")
	}
	query := `
		SELECT
			o.id,
			o.user_id,
			COALESCE(o.email, '') AS email,
			o.currency,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.stripe_payment_intent,
			o.address_id,
			a.addressee,
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
			o.user_id = $1 AND
			o.status = 'pending'
		LIMIT 1
	` // LIMIT 1 added, but technically shouldn't be needed (system should limit users to one pending order)

	// Execute the query
	var rawIntent []byte
	var address struct {
		ID         sql.NullString
		Addressee  sql.NullString
		Line1      sql.NullString
		Line2      sql.NullString
		City       sql.NullString
		State      sql.NullString
		PostalCode sql.NullString
		Country    sql.NullString
	}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&order.ID,
		&order.UserID,
		&order.Email,
		&order.Currency,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&rawIntent,
		&address.ID,
		&address.Addressee,
		&address.Line1,
		&address.Line2,
		&address.City,
		&address.State,
		&address.PostalCode,
		&address.Country,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return order, types.ErrNotFound
	}
	if err != nil {
		return order, err
	}

	// Unmarshal the Stripe payment intent if it exists
	if len(rawIntent) > 0 {
		var payInt stripe.PaymentIntent
		err = json.Unmarshal(rawIntent, &payInt)
		if err != nil {
			return order, fmt.Errorf("failed to unmarshal Stripe payment intent: %w", err)
		}
		order.StripePaymentIntent = &payInt
	}

	// Populate address if it exists
	if address.ID.Valid {
		order.Address = &types.Address{
			ID:         address.ID.String,
			Addressee:  &address.Addressee.String,
			Line1:      address.Line1.String,
			Line2:      &address.Line2.String,
			City:       address.City.String,
			State:      address.State.String,
			PostalCode: address.PostalCode.String,
			Country:    address.Country.String,
		}
	}

	// Populate order items for this order
	if order.Items, err = r.populateOrderItems(ctx, order.ID); err != nil {
		return order, fmt.Errorf("failed to populate order items: %w", err)
	}

	return order, nil
}
