package repositories

import (
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/joho/godotenv"
)

var dbPool *sql.DB

func TestMain(m *testing.M) {
	// Load environment variables from .env
	err := godotenv.Load("../.env")
	if err != nil {
		slog.Warn("No .env file found, using system environment variables")
	}

	dbConfig := types.DBConfig{
		URL:             os.Getenv("DATABASE_URL"),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 5,
		ConnMaxIdleTime: time.Minute * 5,
	}

	dbPool = db.Connect(dbConfig)
	defer dbPool.Close()

	// Initialize ID generator
	utilities.InitIDGenerator(99)

	// Run tests
	code := m.Run()

	dbPool.Close()
	os.Exit(code)
}
