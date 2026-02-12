package logger

import (
	"io"
	"log/slog"
)

func New(loglvlstr string, isProd bool, w io.Writer) *slog.Logger {
	var handler slog.Handler
	lvl := parseLogLevel(loglvlstr)

	if isProd {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl, AddSource: !isProd})
	} else {
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: lvl, AddSource: !isProd})
	}

	return slog.New(handler)
}

func parseLogLevel(lvl string) slog.Level {
	switch lvl {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "error":
		return slog.LevelError
	case "warn", "warning":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
