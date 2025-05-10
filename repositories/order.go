package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
)

type OrderRepository interface {
	GetOrder(ctx context.Context, order *types.Order) error
	GetOrders(ctx context.Context, userID string, page, limit int) ([]types.Order, error)
	PopulateOrderItems(ctx context.Context, orders *[]types.Order) error
	CreateOrder(ctx context.Context, order *types.Order) error
	UpdateOrder(ctx context.Context, order *types.Order) error
	CreateStripeEvent(ctx context.Context, event stripe.Event) error
	CancelPendingOrders(ctx context.Context, interval time.Duration) ([]string, error)
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
		SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
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
			WHERE o.id = oi.order_id AND o.status = 'cancelled'
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

func (r *orderRepository) CreateOrder(ctx context.Context, order *types.Order) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	addr := order.Address
	if err := tx.QueryRowContext(ctx, `
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
			created_at,
			updated_at
		FROM addresses
		WHERE id = $1 AND
			user_id = $2 AND
			is_deleted = FALSE
	`, order.Address.ID, order.UserID).Scan(
		&addr.ID,
		&addr.UserID,
		&addr.Addressee,
		&addr.AddressLine1,
		&addr.AddressLine2,
		&addr.City,
		&addr.StateCode,
		&addr.PostalCode,
		&addr.CountryCode,
		&addr.CreatedAt,
		&addr.UpdatedAt,
	); err != nil {
		return err
	}

	// Retrieve cart items
	var cartItems []types.CartItem
	query := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.tax_code,
			p.images,
			p.description,
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
			&item.Product.Description,
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

	// Calculate cart total (excluding tax + shipping)
	amount := calculateOrderAmount(cartItems)
	if amount == 0 {
		return errors.New("order cart is empty")
	}

	// Reduce inventory
	if err = reduceInventory(ctx, tx, cartItems); err != nil {
		return err
	}

	// create a new order with pending status
	query = `
		INSERT INTO orders (id, user_id, email, address_id, currency, amount) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, email, currency, amount, status, created_at
	`
	if err = tx.QueryRowContext(
		ctx, query,
		order.ID,
		order.UserID,
		order.Email,
		order.Address.ID,
		order.Currency,
		amount,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.Email,
		&order.Currency,
		&order.Amount,
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

// CreateStripeEvent saves a Stripe event to the database
func (r *orderRepository) CreateStripeEvent(ctx context.Context, event stripe.Event) error {
	if event.Data != nil {
		event.Data.Object.ClientSecret = ""
	}
	query := `
		INSERT INTO stripe_events (
			id,
			event_type,
			payload,
			processed_at
		)
		VALUES ($1, $2, $3, $4)
	`
	payload, err := json.Marshal(event.Data.Object)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.Type,
		payload,
		time.Unix(event.Created, 0).UTC(),
	)
	return err
}

// UpdateOrder updates an order with new status and/or payment intent ID
// FIXME break this up into two sep functions
func (r *orderRepository) UpdateOrder(ctx context.Context, order *types.Order) error {
	if order == nil || order.ID == "" {
		return fmt.Errorf("missing order ID")
	}

	query := `UPDATE orders SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}
	argCount := 1

	if order.Status != "" {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, order.Status)
		argCount++
	}

	if order.TaxAmount != 0 {
		query += fmt.Sprintf(", tax_amount = $%d", argCount)
		args = append(args, order.TaxAmount)
		argCount++
	}

	if order.ShippingAmount != 0 {
		query += fmt.Sprintf(", shipping_amount = $%d", argCount)
		args = append(args, order.ShippingAmount)
		argCount++
	}

	if order.TotalAmount != 0 {
		query += fmt.Sprintf(", total_amount = $%d", argCount)
		args = append(args, order.TotalAmount)
		argCount++
	}

	if order.StripePaymentIntent != nil {
		intentJSON, err := json.Marshal(order.StripePaymentIntent)
		if err != nil {
			return fmt.Errorf("failed to encode stripe payment intent: %w", err)
		}
		query += fmt.Sprintf(", stripe_payment_intent = $%d", argCount)
		args = append(args, intentJSON)
		argCount++
	}

	if len(args) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, order.ID)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
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
			a.address_line1,
			a.address_line2,
			a.city,
			a.state_code,
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

	var result []types.Order
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
			&order.Address.AddressLine1,
			&order.Address.AddressLine2,
			&order.Address.City,
			&order.Address.StateCode,
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

	return result, nil
}

// PopulateOrderItems populates the order items for a list of orders
func (r *orderRepository) PopulateOrderItems(ctx context.Context, orders *[]types.Order) error {
	if len(*orders) == 0 {
		return nil
	}

	// Collect order IDs
	orderIDs := make([]interface{}, len(*orders))
	for i, order := range *orders {
		orderIDs[i] = order.ID
	}

	// Dynamically build the query with placeholders
	placeholders := make([]string, len(orderIDs))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1) // PostgreSQL uses $1, $2, ...
	}

	// Query to fetch order items
	query := fmt.Sprintf(`
		SELECT
			order_id,
			product_id,
			description,
			thumbnail,
			quantity,
			unit_price
		FROM v_order_items
		WHERE order_id IN (%s)
	`, strings.Join(placeholders, ","))

	// Query to fetch order items
	rows, err := r.db.QueryContext(ctx, query, orderIDs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Map to store items grouped by order ID
	itemMap := make(map[string][]types.OrderItem)

	// Process query results
	for rows.Next() {
		var orderID string
		item := types.OrderItem{}
		item.Product = types.Product{}
		if err := rows.Scan(
			&orderID,
			&item.Product.ID,
			&item.Product.Description,
			&item.Thumbnail,
			&item.Quantity,
			&item.UnitPrice,
		); err != nil {
			return err
		}
		itemMap[orderID] = append(itemMap[orderID], item)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return err
	}

	// Populate the orders with their items
	for i, order := range *orders {
		if items, ok := itemMap[order.ID]; ok {
			(*orders)[i].Items = items
		}
	}

	return nil
}

// isEmptyOrderLookup checks if the order lookup is empty
func isEmptyOrderLookup(order *types.Order) bool {
	return order.ID == "" && (order.StripePaymentIntent == nil || order.StripePaymentIntent.ID == "")
}

// GetOrder retrieves an order by ID or payment intent ID
// FIXME refactor split into two functions,
// GetOrderByID
// GetOrderByPaymentIntent
func (r *orderRepository) GetOrder(ctx context.Context, order *types.Order) error {
	// Validate input
	if isEmptyOrderLookup(order) {
		return fmt.Errorf("missing identifier: provide order.ID or StripePaymentIntent.ID")
	}
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
			a.address_line1,
			a.address_line2,
			a.city,
			a.state_code,
			a.postal_code,
			o.created_at,
			o.updated_at
		FROM orders o
		JOIN addresses a ON o.address_id = a.id
	`
	args := []interface{}{}
	var whereClause string

	// Build the WHERE clause based on provided fields
	if order.ID != "" {
		whereClause = "WHERE o.id = $1"
		args = append(args, order.ID)
	} else if order.UserID != "" {
		whereClause = "WHERE o.user_id = $1 ORDER BY o.created_at DESC LIMIT 1"
		args = append(args, order.UserID)
	} else if order.StripePaymentIntent != nil && order.StripePaymentIntent.ID != "" {
		whereClause = "WHERE o.stripe_payment_intent->>'id' = $1"
		args = append(args, order.StripePaymentIntent.ID)
	}

	// Combine query and where clause
	query += whereClause

	// Execute the query
	order.Address = &types.Address{}
	order.StripePaymentIntent = &stripe.PaymentIntent{} // Avoid overwriting the existing pointer
	var rawIntent []byte
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
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
		&order.Address.AddressLine1,
		&order.Address.AddressLine2,
		&order.Address.City,
		&order.Address.StateCode,
		&order.Address.PostalCode,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("order not found")
		}
		return err
	}

	// Unmarshal the Stripe payment intent if it exists
	if len(rawIntent) > 0 {
		var spi stripe.PaymentIntent
		err = json.Unmarshal(rawIntent, &spi)
		if err != nil {
			return fmt.Errorf("failed to unmarshal Stripe payment intent: %w", err)
		}
		order.StripePaymentIntent = &spi
	}

	return nil
}
