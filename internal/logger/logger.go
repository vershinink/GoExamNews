package logger

import (
	"log/slog"
	"os"
)

// MustLoad - инициализирует логгер из пакета slog.
func MustLoad() *slog.Logger {
	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	return log
}

// Err - обертка для ошибки, представляет ее как атрибут слоггера.
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
