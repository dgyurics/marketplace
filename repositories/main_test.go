package repositories

import (
	"database/sql"
	"os"
	"testing"

	"github.com/dgyurics/marketplace/db"
)

var dbPool *sql.DB

func TestMain(m *testing.M) {
	dbURL := "postgres://postgres:postgres@localhost:5432/marketplace?sslmode=disable"
	os.Setenv("DATABASE_URL", dbURL)

	dbPool = db.Connect()
	defer dbPool.Close()

	// Run tests
	code := m.Run()

	dbPool.Close()
	os.Exit(code)
}
