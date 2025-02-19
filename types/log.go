package types

import "log/slog"

type LoggerConfig struct {
	LogFilePath string
	AppID       string
	Level       slog.Level
}
