package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dgyurics/marketplace/types"
	_ "github.com/lib/pq"
)

// Connect creates a connection to the database with retry logic
func Connect(c types.DBConfig) *sql.DB {
	var db *sql.DB
	var err error
	maxRetries := 10              // TODO make configurable
	retryDelay := 5 * time.Second // TODO make configurable
	dataSourceName := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dataSourceName)
		if err == nil {
			if err = db.Ping(); err == nil {
				break // Successfully connected
			}
		}

		if i < maxRetries-1 {
			slog.Warn("Failed to connect to database, retrying...", "attempt", i+1, "error", err)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		slog.Error("Failed to connect to database", "host", c.Host, "port", c.Port, "user", c.User, "db", c.Name, "sslmode", c.SSLMode, "error", err)
		os.Exit(1)
	}

	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)

	slog.Info("Connected to database")
	return db
}

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB) error {
	// create migrations table, if not exists
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// TODO
	// get applied migrations
	// get list of migration files (returns in sorted ascending order)
	// iterate migration files list,
	// if filename not exist in applied migrations, exec runMigration

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		)`
	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	applied := make(map[string]bool)
	return applied, nil
}

func runMigration(db *sql.DB, filename string) error {
	// read the file
	// split commands by semicolon
	// run each command
	return nil
}

func getMigrationFiles() ([]string, error) {
	var files []string
	// retrieve filenames from db/migrations directory
	// sort in ascending order
	return files, nil
}
