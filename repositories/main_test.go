package repositories

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/dgyurics/marketplace/db"
)

var dbPool *sql.DB

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/marketplace?sslmode=disable"
	}

	var err error
	dbPool, err = db.Connect(dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// Run tests
	code := m.Run()

	dbPool.Close()
	os.Exit(code)
}
