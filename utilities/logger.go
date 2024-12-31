package utilities

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var logFile *os.File // Keep a reference to the log file to close it later

// InitLogger initializes the logger with the given configuration.
func InitLogger(config LoggerConfig) {
	logFile, openErr := os.OpenFile(config.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		fmt.Printf("Failed to open log file: %v\n", openErr)
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})
		slog.SetDefault(slog.New(handler).With("fallback", true))
	}

	handler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: config.Level})
	log := slog.New(handler).WithGroup("app").With("id", config.AppID)
	slog.SetDefault(log)
}

// CloseLogger closes the log file.
func CloseLogger() {
	if logFile != nil {
		slog.Info("Closing log file")
		logFile.Close()
	}
}

type LoggerConfig struct {
	LogFilePath string
	AppID       string
	Level       slog.Level
}

// LoadLoggerConfig loads logger configuration from environment variables.
func LoadLoggerConfig() LoggerConfig {
	levelStr := GetEnv("LOG_LEVEL") // returns string
	level := parseLogLevel(levelStr)
	return LoggerConfig{
		LogFilePath: GetEnv("LOG_FILE_PATH"),
		AppID:       GetEnv("APP_ID"),
		Level:       level,
	}
}

// parseLogLevel converts a string log level to slog.Level.
func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default to Info if the level is not recognized
	}
}
