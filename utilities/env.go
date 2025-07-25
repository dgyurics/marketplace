package utilities

import (
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
		"./deploy/env/.env.dev",     // When run from project root
		"../deploy/env/.env.dev",    // When run from a direct subdirectory
		"../../deploy/env/.env.dev", // When run from a package in a subdirectory
	}

	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			return // Successfully loaded
		}
	}

	panic("Error loading environment variables: .env.dev not found")
}
