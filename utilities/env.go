package utilities

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func LoadEnvironment() {
	env, _ := os.LookupEnv("ENVIRONMENT")
	if strings.EqualFold(env, "production") {
		return
	}

	// Try multiple possible locations
	envPaths := []string{
		"./deploy/local/.env",
		"../deploy/local/.env",
		"../../deploy/local/.env",
	}

	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			return // Successfully loaded
		}
	}

	slog.Error("Error loading environment variables: .env not found")
	os.Exit(1)
}
