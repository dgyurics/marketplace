// logger utility initializes the structured logger, slog.
// To access the logger throughout the application,
// simply reference the slog package "log/slog"
package utilities

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/dgyurics/marketplace/types"
)

var logFile *os.File // Keep a reference to the log file to close it later

// InitLogger initializes the logger with the given configuration.
func InitLogger(config types.LoggerConfig) {
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

// loadLoggerConfig loads logger configuration from environment variables.
func loadLoggerConfig() types.LoggerConfig {
	return types.LoggerConfig{
		LogFilePath: mustLookupEnv("LOG_FILE_PATH"),
		AppID:       mustLookupEnv("APP_ID"),
		Level:       parseLogLevel(mustLookupEnv("LOG_LEVEL")),
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

// ErrorLog adapts slog to implement the log.Logger interface.
//
// By implementing the Write method, ErrorLog makes it possible
// to pass slog as the logger for http.Server, which requires
// a logger conforming to the log.Logger interface.
type ErrorLog struct{}

func (s *ErrorLog) Write(p []byte) (n int, err error) {
	slog.Error(string(p))
	return len(p), nil
}
