package services

import (
	"os"
	"testing"

	"github.com/dgyurics/marketplace/utilities"
)

func TestMain(m *testing.M) {
	// Initialize ID generator
	utilities.InitIDGenerator(99)

	// Run tests
	code := m.Run()

	os.Exit(code)
}
