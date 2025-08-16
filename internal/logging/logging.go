package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/vinegarhq/vinegar/internal/dirs"
)

const (
	LevelWine   = slog.LevelInfo + 1
	LevelRoblox = slog.LevelInfo + 2
)

func NewTextHandler(w io.Writer, noColor bool) slog.Handler {
	return tint.NewHandler(w, &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.TimeOnly,
		NoColor:    noColor,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.LevelKey || len(groups) != 0 {
				return a
			}
			switch a.Value.Any().(slog.Level) {
			case LevelWine:
				return tint.Attr(1, slog.String(a.Key, "WIN"))
			case LevelRoblox:
				return tint.Attr(6, slog.String(a.Key, "RBX"))
			}
			return a
		},
	})
}

func NewFile() (*os.File, error) {
	if err := dirs.Mkdirs(dirs.Logs); err != nil {
		return nil, err
	}

	// name-2006-01-02T15:04:05Z07:00.log
	path := filepath.Join(dirs.Logs, time.Now().Format(time.RFC3339)+".log")

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return file, nil
}
