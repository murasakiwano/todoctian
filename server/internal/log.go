package internal

import (
	"log/slog"
	"os"
)

func NewLogger(caller string) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}).WithAttrs([]slog.Attr{slog.String("caller", caller)}))
}
