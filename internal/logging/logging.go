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
	LevelDebug   = slog.LevelDebug
	LevelInfo    = slog.LevelInfo
	LevelWine    = slog.Level(2)
	LevelRoblox  = slog.Level(3)
	LevelWarning = slog.LevelWarn
	LevelError   = slog.LevelError
)

// ANSI modes inherited from tint
const (
	ansiReset        = "\033[0m"
	ansiRed          = "\033[31m"
	ansiCyan         = "\033[37m"
	ansiBrightRed    = "\033[91m"
	ansiBrightGreen  = "\033[92m"
	ansiBrightYellow = "\033[93m"
)

func NewTextHandler(w io.Writer, noColor bool) slog.Handler {
	return tint.NewHandler(w, &tint.Options{
		Level:      LevelInfo,
		TimeFormat: time.TimeOnly,
		NoColor:    noColor,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				value := "ERROR"
				color := ""
				switch {
				case level < LevelInfo:
					value = "DBUG"
				case level < LevelWine:
					value = "INFO"
					color = ansiBrightGreen
				case level < LevelRoblox:
					value = "WINE"
					color = ansiRed
				case level < LevelWarning:
					value = "RBLX"
					color = ansiCyan
				case level < LevelError:
					value = "WARN"
					color = ansiBrightYellow
				default:
					value = "ERRO"
					color = ansiBrightRed
				}
				if noColor {
					a.Value = slog.StringValue(value)
				} else {
					a.Value = slog.StringValue(color + value + ansiReset)
				}
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
