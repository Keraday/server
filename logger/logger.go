package logger

import (
	"io"
	"log/slog"
)

var lvl = &slog.LevelVar{}

func Init(debug bool, out io.Writer) {
	if debug {
		lvl.Set(slog.LevelDebug)
	} else {
		lvl.Set(slog.LevelInfo)
	}
	handler := slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: lvl,
		// TODO изменить формат
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Debug("Debug mode")
}
