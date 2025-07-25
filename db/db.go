package db

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/dgyurics/marketplace/types"
	_ "github.com/lib/pq"
)

// Connect creates a connection to the database with retry logic
func Connect(c types.DBConfig) *sql.DB {
	var db *sql.DB
	var err error
	maxRetries := 10
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", c.URL)
		if err == nil {
			if err = db.Ping(); err == nil {
				break // Successfully connected
			}
		}

		if i < maxRetries-1 {
			slog.Info("Failed to connect to database, retrying...", "attempt", i+1, "error", err)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		panic("Failed to connect to database after retries: " + err.Error())
	}

	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)

	slog.Info("Connected to database")
	return db
}
