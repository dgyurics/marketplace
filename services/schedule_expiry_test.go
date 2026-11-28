package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type expiryPaymentStub struct {
	PaymentService
	err   error
	calls int
}

func (s *expiryPaymentStub) CancelPaymentIntent(_ context.Context, _ string) error {
	s.calls++
	return s.err
}

func TestCancelStaleOrder(t *testing.T) {
	env, err := godotenv.Read("../deploy/local/.env")
	if err != nil {
		t.Skip("local PostgreSQL configuration unavailable")
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", env["POSTGRES_HOST"], env["POSTGRES_PORT"], env["POSTGRES_USER"], env["POSTGRES_PASSWORD"], env["POSTGRES_DB"], env["POSTGRES_SSLMODE"])
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)
	require.NoError(t, db.Ping())
	_, err = db.Exec(`CREATE TEMP TABLE orders (id text PRIMARY KEY, status text, payment_method text, payment_reference text, updated_at timestamp, user_id text, tax_amount bigint, total_amount bigint, payment_status text);
 CREATE TEMP TABLE products (id text PRIMARY KEY, inventory integer, updated_at timestamp);
 CREATE TEMP TABLE order_items (order_id text, product_id text, quantity integer)`)
	require.NoError(t, err)
	cases := []struct {
		name, status         string
		reference            any
		paymentErr           error
		wantStatus           string
		wantInventory, calls int
		wantError            bool
	}{
		{"draft", "draft", nil, nil, "canceled", 5, 0, false},
		{"pending canceled at provider", "pending", "pi_test", nil, "canceled", 5, 1, false},
		{"pending processing or provider error", "pending", "pi_test", errors.New("not canceled"), "pending", 3, 1, true},
		{"pending missing reference", "pending", nil, nil, "pending", 3, 0, true},
		{"already paid", "paid", "pi_test", nil, "paid", 3, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.Exec(`TRUNCATE orders, products, order_items;
   INSERT INTO products VALUES ('p', 3, NOW());
   INSERT INTO order_items VALUES ('o', 'p', 2)`)
			require.NoError(t, err)
			_, err = db.Exec(`INSERT INTO orders (id, status, payment_method, payment_reference, updated_at) VALUES ('o', $1, 'stripe', $2, NOW() - INTERVAL '1 hour')`, tc.status, tc.reference)
			require.NoError(t, err)
			payment := &expiryPaymentStub{err: tc.paymentErr}
			scheduler := &scheduleService{db: db, payment: payment}
			err = scheduler.cancelStaleOrder(context.Background(), "o")
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			var status string
			var inventory int
			require.NoError(t, db.QueryRow(`SELECT status FROM orders WHERE id = 'o'`).Scan(&status))
			require.NoError(t, db.QueryRow(`SELECT inventory FROM products WHERE id = 'p'`).Scan(&inventory))
			require.Equal(t, tc.wantStatus, status)
			require.Equal(t, tc.wantInventory, inventory)
			require.Equal(t, tc.calls, payment.calls)
		})
	}

}
