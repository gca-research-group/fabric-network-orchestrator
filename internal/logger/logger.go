package logger

import (
	"log/slog"
	"os"
)

func Setup() error {
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	slog.SetDefault(slog.New(consoleHandler))

	return nil
}
