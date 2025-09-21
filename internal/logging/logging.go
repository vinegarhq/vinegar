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

// Path to the log file that is scoped to the entire program runtime
var Path string

// Level is a custom type to represent custom log levels with
// their names.
type Level int

const (
	LevelWine   = Level(slog.LevelInfo + 1)
	LevelRoblox = Level(slog.LevelInfo + 2)
)

type Handler struct {
	slog.Handler
	file slog.Handler

	// Channel that can be opened & closed to read incomding
	// log records, used to show new logs as it comes.
	ReadRecord func(*slog.Record)
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
	if err := dirs.Mkdirs(dirs.Logs); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(Path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func NewHandler(w io.Writer, level slog.Level) slog.Handler {
	h := NewTextHandler(w, level, true)

	var fh slog.Handler
	var r slog.Record
	f, err := openPath()
	if err == nil {
		fh = NewTextHandler(f, level, false)
		r = slog.NewRecord(time.Now(), slog.LevelInfo, "Logging to file", 0)
		r.AddAttrs(slog.String("path", Path))
	} else {
		r = slog.NewRecord(time.Now(), slog.LevelError, "Failed to open log file", 0)
		r.AddAttrs(slog.String("err", err.Error()))
	}
	h.Handle(context.TODO(), r)

	return &Handler{
		Handler: h,
		file:    fh,
	}
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
	if h.ReadRecord != nil {
		h.ReadRecord(&r)
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

func NewTextHandler(w io.Writer, level slog.Level, color bool) slog.Handler {
	return tint.NewHandler(w, &tint.Options{
		Level:      level,
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
