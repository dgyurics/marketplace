package db

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/dgyurics/marketplace/utilities"
	_ "github.com/lib/pq"
)

// Connect creates a connection to the database and returns the connection.
// The returned [sql.DB] is safe for concurrent use by multiple goroutines,
// and should only be called once.
func Connect() *sql.DB {
	databaseURL := utilities.GetEnv("DATABASE_URL")
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to database")
	return db
}
