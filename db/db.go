package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
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

	// get migration filenames (ascending order)
	filenames, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to retrieve migration files: %w", err)
	}

	// get applied migration filenames (in map form)
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to retrieve applied migrations: %w", err)
	}

	// if filename not exist in applied migrations, exec runMigration
	for _, filename := range filenames {
		if _, ok := appliedMigrations[filename]; !ok {
			if err := runMigration(db, filename); err != nil {
				return fmt.Errorf("failed to run migration file %s: %w", filename, err)
			}
			slog.Info("Applied migration", "filename", filename)
		}
	}

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

// getAppliedMigrations returns a map of applied migrations
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := db.Query(`
		SELECT filename
		FROM schema_migrations
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}

	return applied, nil
}

// runMigration retrieves the file filename and executes the SQL statements within it
// within a transaction, recording the migration in the schema_migrations table upon success
func runMigration(db *sql.DB, filename string) error {
	content, err := os.ReadFile("db/migrations/" + filename)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(content)); err != nil {
		return err
	}

	// Record the migration
	if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", filename); err != nil {
		return err
	}

	return tx.Commit()
}

// getMigrationFiles returns an array of migration filenames in ascending order
func getMigrationFiles() ([]string, error) {
	entries, err := os.ReadDir("db/migrations")
	if err != nil {
		return nil, err
	}

	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			filenames = append(filenames, entry.Name())
		}
	}

	return filenames, nil
}
