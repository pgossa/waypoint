package ui

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func InitLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

func Log() *slog.Logger {
	return logger
}
