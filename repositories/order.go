package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgyurics/marketplace/types"
)

type OrderRepository interface {
	GetOrder(ctx context.Context, order *types.Order) error
	GetOrders(ctx context.Context, userID string, page, limit int) ([]types.Order, error)
	PopulateOrderItems(ctx context.Context, orders *[]types.Order) error
	CreateOrder(ctx context.Context, userID, addressID string) (types.Order, error)
	UpdateOrder(ctx context.Context, order *types.Order) error
	CreateStripeEvent(ctx context.Context, event types.StripeEvent) error
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

// CreateOrder creates a new order from the user's cart
// TODO - implement shipping and tax calculation, possibly as a separate function
// Most likely will need to implement update_or_create_order_from_cart in Go
// if planning to move to different database as well...
func (r *orderRepository) CreateOrder(ctx context.Context, userID, addressID string) (order types.Order, err error) {
	// 1) Create or update the order from the user's cart
	query := "SELECT update_or_create_order_from_cart($1, $2)"
	var orderID string
	err = r.db.QueryRowContext(ctx, query, userID, addressID).Scan(&orderID)
	if err != nil {
		return order, err
	}

	// 2) Retrieve the new or updated order
	query = `
	  SELECT
			o.id,
			o.user_id,
			o.currency,
			o.amount,
			o.tax_amount,
			o.shipping_amount,
			o.total_amount,
			o.status,
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
		WHERE o.id = $1
	`
	// Execute the query
	order.Address = &types.Address{}
	err = r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Currency,
		&order.Amount,
		&order.TaxAmount,
		&order.ShippingAmount,
		&order.TotalAmount,
		&order.Status,
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

	return order, err
}

// CreateStripeEvent saves a Stripe event to the database
func (r *orderRepository) CreateStripeEvent(ctx context.Context, event types.StripeEvent) error {
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
func (r *orderRepository) UpdateOrder(ctx context.Context, order *types.Order) error {
	if order.ID == "" {
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
			var spi types.StripePaymentIntent
			if err := json.Unmarshal(rawIntent, &spi); err != nil {
				return nil, fmt.Errorf("failed to unmarshal Stripe payment intent: %w", err)
			}
			order.StripePaymentIntent = &spi
		}

		result = append(result, order)
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
		if err := rows.Scan(
			&orderID,
			&item.ProductID,
			&item.Description,
			&item.Thumbnail,
			&item.Quantity,
			&item.UnitPrice,
		); err != nil {
			return err
		}
		itemMap[orderID] = append(itemMap[orderID], item)
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
	return order.ID == "" && order.UserID == "" && (order.StripePaymentIntent == nil || order.StripePaymentIntent.ID == "")
}

// GetOrder retrieves an order by ID, user ID, or payment intent ID
func (r *orderRepository) GetOrder(ctx context.Context, order *types.Order) error {
	// Validate input
	if isEmptyOrderLookup(order) {
		return fmt.Errorf("missing identifier: provide order.ID, order.UserID, or StripePaymentIntent.ID")
	}

	query := `
		SELECT
			o.id,
			o.user_id,
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
	order.StripePaymentIntent = &types.StripePaymentIntent{} // Avoid overwriting the existing pointer
	var rawIntent []byte
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&order.ID,
		&order.UserID,
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
		var spi types.StripePaymentIntent
		err = json.Unmarshal(rawIntent, &spi)
		if err != nil {
			return fmt.Errorf("failed to unmarshal Stripe payment intent: %w", err)
		}
		order.StripePaymentIntent = &spi
	}

	return nil
}
