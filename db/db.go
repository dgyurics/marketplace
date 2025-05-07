package db

import (
	"database/sql"
	"log/slog"

	"github.com/dgyurics/marketplace/types"
	_ "github.com/lib/pq"
)

// Connect creates a connection to the database and returns the connection.
// The returned [sql.DB] is safe for concurrent use by multiple goroutines,
// and should only be called once.
func Connect(c types.DBConfig) *sql.DB {
	db, err := sql.Open("postgres", c.URL)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	if err := db.Ping(); err != nil {
		panic("Failed to ping database: " + err.Error())
	}

	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)

	slog.Info("Connected to database")
	return db
}
