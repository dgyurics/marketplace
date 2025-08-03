// logger utility initializes the structured logger, slog.
// To access the logger throughout the application,
// simply reference the slog package "log/slog"
package utilities

import (
	"log/slog"
	"os"
	"strings"

	"github.com/dgyurics/marketplace/types"
)

// InitLogger initializes the logger with the given log level.
func InitLogger(config types.LoggerConfig) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: config.Level})
	slog.SetDefault(slog.New(handler))
}

// loadLoggerConfig loads logger configuration from environment variables.
func loadLoggerConfig() types.LoggerConfig {
	return types.LoggerConfig{
		Level: parseLogLevel(mustLookupEnv("LOG_LEVEL")),
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
