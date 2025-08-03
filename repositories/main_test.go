package repositories

import (
	"database/sql"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

var dbPool *sql.DB

func TestMain(m *testing.M) {
	utilities.InitLogger(types.LoggerConfig{
		Level: slog.LevelDebug,
	})
	// Load environment variables
	utilities.LoadEnvironment()
	port, _ := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	dbConfig := types.DBConfig{
		Host:            os.Getenv("POSTGRES_HOST"),
		Port:            port,
		User:            os.Getenv("POSTGRES_USER"),
		Password:        os.Getenv("POSTGRES_PASSWORD"),
		Name:            os.Getenv("POSTGRES_NAME"),
		SSLMode:         os.Getenv("POSTGRES_SSLMODE"),
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
