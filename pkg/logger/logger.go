package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func NewLogger() *Logger {
	opts := slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}

	h := slog.NewTextHandler(os.Stdout, &opts)
	newLogger := slog.New(h)
	return &Logger{newLogger}
}
