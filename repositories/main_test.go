package repositories

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/models"
)

var dbPool *sql.DB

func TestMain(m *testing.M) {
	dbConfig := models.DBConfig{
		URL:             "postgres://postgres:postgres@localhost:5432/marketplace?sslmode=disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 5,
		ConnMaxIdleTime: time.Minute * 5,
	}

	dbPool = db.Connect(dbConfig)
	defer dbPool.Close()

	// Run tests
	code := m.Run()

	dbPool.Close()
	os.Exit(code)
}
