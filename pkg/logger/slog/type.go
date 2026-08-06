package slog

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	Level  string
	Writer io.Writer
}

type Logger = slog.Logger

func New(cfg *Config) *Logger {

	lvl := LevelInfo
	switch strings.ToUpper(cfg.Level) {
	case "DEBUG":
		lvl = LevelDebug

	case "WARN":
		lvl = LevelWarn

	case "ERROR":
		lvl = LevelError
	}

	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	handler := slog.NewJSONHandler(
		cfg.Writer, &slog.HandlerOptions{Level: lvl, ReplaceAttr: replaceAttr},
	)
	return slog.New(handler)
}

func replaceAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.TimeKey && attr.Value.Kind() == slog.KindTime {
		return epochTimeEncoder("ts", attr.Value.Time())
	}
	return attr
}

func epochTimeEncoder(key string, t time.Time) slog.Attr {
	return slog.Attr{
		Key: key, Value: slog.Float64Value(float64(t.UnixNano()) / float64(time.Second)),
	}
}
