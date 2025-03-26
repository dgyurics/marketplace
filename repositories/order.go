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
	CreateOrder(ctx context.Context, userID, addressID string) (*types.Order, error)
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
func (r *orderRepository) CreateOrder(ctx context.Context, userID, addressID string) (*types.Order, error) {
	// 1) Create or update the order from the user's cart
	query := "SELECT update_or_create_order_from_cart($1, $2)"
	var orderID string
	err := r.db.QueryRowContext(ctx, query, userID, addressID).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	// 2) Retrieve the new or updated order
	query = `
	  SELECT
			id,
			user_id,
			currency,
			amount,
			tax_amount,
			shipping_amount,
			total_amount,
			status,
			payment_intent_id,
			address_id,
			created_at,
			updated_at
		FROM orders
		WHERE id = $1
	`
	order := &types.Order{}
	err = r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Currency,
		&order.Amount,
		&order.TaxAmount,
		&order.ShippingAmount,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentIntentID,
		&order.AddressID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order: %w", err)
	}
	return order, nil
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

	query := `
		UPDATE orders
		SET updated_at = CURRENT_TIMESTAMP
	`
	args := []interface{}{}
	argCount := 1

	if order.Status != "" {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, order.Status)
		argCount++
	}

	if order.PaymentIntentID != "" {
		query += fmt.Sprintf(", payment_intent_id = $%d", argCount)
		args = append(args, order.PaymentIntentID)
		argCount++
	}

	// Ensure there's something to update
	if len(args) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Add the WHERE clause
	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, order.ID)

	// Execute the query
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetOrders retrieves all orders for a user
func (r *orderRepository) GetOrders(ctx context.Context, userID string, page, limit int) ([]types.Order, error) {
	query := `
		SELECT
			id,
			user_id,
			currency,
			amount,
			tax_amount,
			total_amount,
			status,
			payment_intent_id,
			address_id,
			created_at,
			updated_at
		FROM orders
		WHERE user_id = $1 AND status != 'pending'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []types.Order{}
	for rows.Next() {
		order := types.Order{}
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Currency,
			&order.Amount,
			&order.TaxAmount,
			&order.TotalAmount,
			&order.Status,
			&order.PaymentIntentID,
			&order.AddressID,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
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

// GetOrder retrieves an order by ID, user ID, or payment intent ID
func (r *orderRepository) GetOrder(ctx context.Context, order *types.Order) error {
	// Validate input
	if order.ID == "" && order.UserID == "" && order.PaymentIntentID == "" {
		return fmt.Errorf("at least one of order.ID, order.UserID, or order.PaymentIntentID must be provided")
	}

	query := `
		SELECT
			id,
			user_id,
			currency,
			amount,
			tax_amount,
			total_amount,
			status,
			payment_intent_id,
			address_id,
			created_at,
			updated_at
		FROM orders
	`
	args := []interface{}{}
	var whereClause string

	// Build the WHERE clause based on provided fields
	if order.ID != "" {
		whereClause = "WHERE id = $1"
		args = append(args, order.ID)
	} else if order.UserID != "" {
		whereClause = "WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1"
		args = append(args, order.UserID)
	} else if order.PaymentIntentID != "" {
		whereClause = "WHERE payment_intent_id = $1"
		args = append(args, order.PaymentIntentID)
	}

	// Combine query and where clause
	query += whereClause

	// Execute the query
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&order.ID,
		&order.UserID,
		&order.Currency,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentIntentID,
		&order.AddressID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("order not found")
		}
		return err
	}

	return nil
}
