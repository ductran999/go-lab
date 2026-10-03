// Package logx is a slog Handler painting levels in ANSI colors for
// local dev: DEBUG grey, INFO green, WARN yellow, ERROR red.
// Production stays JSON (slog.NewJSONHandler) — colors die in files.
package logx

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

const (
	grey   = "\033[90m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	reset  = "\033[0m"
)

// New returns a colorizing handler writing to w (os.Stdout).
// LOG_COLOR=0 disables (pipes, CI).
func New(w io.Writer) slog.Handler {
	if os.Getenv("LOG_COLOR") == "0" {
		return slog.NewTextHandler(w, nil)
	}

	return &handler{w: w}
}

type handler struct {
	mu sync.Mutex
	w  io.Writer
}

func colorOf(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return red
	case level >= slog.LevelWarn:
		return yellow
	case level >= slog.LevelInfo:
		return green
	default:
		return grey
	}
}

func (h *handler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := fmt.Fprintf(h.w, "%s%s %-5s%s %s",
		grey, r.Time.Format("15:04:05"), colorOf(r.Level)+r.Level.String()+grey, reset, r.Message)
	if err != nil {
		return err
	}

	r.Attrs(func(a slog.Attr) bool {
		_, err := fmt.Fprintf(h.w, " %s=%v", a.Key, a.Value)

		return err == nil
	})

	_, err = fmt.Fprintln(h.w, reset)

	return err
}

func (h *handler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

func (h *handler) WithGroup(_ string) slog.Handler { return h }
