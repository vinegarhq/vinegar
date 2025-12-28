// Logging implements levels and a custom handler for use in Vinegar.
package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/vinegarhq/vinegar/internal/dirs"
)

// Path to the log file that is scoped to the entire program runtime.
var Path string

// Level is a custom type to represent custom log levels with
// their names.
type Level int

// LoggerLevel is the level of the default Handler.
var LoggerLevel slog.Level

const (
	LevelWine   = Level(slog.LevelInfo + 1)
	LevelRoblox = Level(slog.LevelInfo + 2)
)

// Handler is a slog handler with additional extra level types
// that outputs to a file, determined by Path, set at package init.
type Handler struct {
	slog.Handler
	file slog.Handler
}

func init() {
	// name-2006-01-02T15:04:05Z07:00.log
	Path = filepath.Join(dirs.Logs, time.Now().Format(time.RFC3339)+".log")
}

// Level implements slog.Leveler.
func (l Level) Level() slog.Level {
	return slog.Level(l)
}

// FromLevel is a helper to transform slog.Level to Level
func FromLevel(l slog.Level) Level {
	return Level(int(l))
}

// String implements Stringer, used to represent the level's name
// in [Handler].
func (l Level) String() string {
	switch l {
	case LevelWine:
		return "WIN"
	case LevelRoblox:
		return "RBX"
	default:
		return l.Level().String()
	}
}

func openPath() (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(Path), 0o755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(Path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func cleanupLogs() error {
	dir := filepath.Dir(Path)
	logs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Immediately start removing log files that aren't beyond
	// minimum count and this file is not managed by Vinegar or is old
	// (determined by filename length)
	for i, log := range logs {
		if i > len(logs)-16 || len(log.Name()) != 29 {
			continue
		}
		if err := os.Remove(filepath.Join(dir, log.Name())); err != nil {
			return err
		}
	}

	return nil
}

// NewHandler creates a new [Handler] that writes to w. When called,
// the same handler will be used to notify the current log filepath,
// and old logs in its directory will also be removed and printed.
func NewHandler(w io.Writer, level slog.Level) slog.Handler {
	h := &Handler{Handler: NewTextHandler(w, true)}
	l := slog.New(h)

	f, err := openPath()
	if err == nil {
		h.file = NewTextHandler(f, false)
		l.Info("Logging to file", "path", Path)
	} else {
		l.Error("Failed to open log file", "err", err)
	}

	if err := cleanupLogs(); err != nil {
		l.Error("Failed to clear old log files", "err", err)
	}

	return h
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	c := r.Clone()
	herr := h.Handler.Handle(ctx, c)
	var ferr error
	if h.file != nil {
		ferr = h.file.Handle(ctx, r)
	}
	if herr != nil || ferr != nil {
		return errors.Join(herr, ferr)
	}
	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var a slog.Handler
	if h.file != nil {
		a = h.file.WithAttrs(attrs)
	}
	return &Handler{
		Handler: h.Handler.WithAttrs(attrs),
		file:    a,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	var g slog.Handler
	if h.file != nil {
		g = h.file.WithGroup(name)
	}
	return &Handler{
		Handler: h.Handler.WithGroup(name),
		file:    g,
	}
}

// NewTextHandler is a wrapper for [tint.NewHandler] for handling the custom log
// levels exported in this package.
func NewTextHandler(w io.Writer, color bool) slog.Handler {
	return tint.NewHandler(w, &tint.Options{
		Level:      &LoggerLevel,
		TimeFormat: time.TimeOnly,
		NoColor:    !color,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.LevelKey || len(groups) != 0 {
				return a
			}
			switch l := FromLevel(a.Value.Any().(slog.Level)); l {
			case LevelWine:
				return tint.Attr(1, slog.String(a.Key, l.String()))
			case LevelRoblox:
				return tint.Attr(6, slog.String(a.Key, l.String()))
			}
			return a
		},
	})
}
