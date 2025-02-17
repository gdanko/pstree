package logger

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

var (
	Logger *slog.Logger
	once   sync.Once
)

type CustomHandler struct {
	level slog.Level
}

func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	fmt.Printf("[%s] %s\n", r.Level, r.Message)
	return nil
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}

func Init(level slog.Level) {
	once.Do(func() {
		Logger = slog.New(&CustomHandler{level: level})
	})
}
