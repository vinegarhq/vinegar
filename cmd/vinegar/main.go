package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/pojntfx/go-gettext/pkg/i18n"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/sysinfo"
)

var LocaleDir = "/usr/share/locale"

func main() {
	debug.SetPanicOnFault(true)

	if slices.ContainsFunc(sysinfo.Cards, func(c sysinfo.Card) bool {
		return strings.HasPrefix(c.Driver, "nvidia") &&
			os.Getenv("WAYLAND_DISPLAY") != "" // GDK will prefer wayland
	}) {
		// Causes GTK panic with: "No provider of eglGetCurrentContext found.""
		os.Setenv("GSK_RENDERER", "cairo")
	}
	// VK_SUBOPTIMAL_KHR
	os.Setenv("GDK_DISABLE", "vulkan")

	if err := i18n.InitI18n("vinegar", LocaleDir, slog.Default()); err != nil {
		slog.Error("Failed to set locale", "err", err)
	}

	// Handle mcp before GApplication so that this process retains the
	// caller's stdin/stdout — GLib would otherwise forward the command line
	// to the primary instance, which has the wrong stdio for MCP.
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "run" {
		args = args[1:]
	}
	if len(args) == 1 && args[0] == "mcp" {
		os.Exit(runMCP())
	}

	if code := newApp().Run(int32(len(os.Args)), os.Args); code > 0 {
		os.Exit(int(code))
	}
}

func runMCP() int {
	slog.SetDefault(slog.New(logging.NewHandler(os.Stderr, slog.LevelInfo)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "err", err)
		return 1
	}

	pfx := cfg.Prefix()
	if !pfx.Exists() {
		slog.Error("Wine prefix does not exist, please install Studio first")
		return 1
	}
	if !pfx.Running() {
		slog.Error("Studio is not running")
		return 1
	}

	entries, err := os.ReadDir(dirs.Versions)
	if err != nil {
		slog.Error("Failed to read versions directory", "err", err)
		return 1
	}

	var mcpPath string
	for _, e := range entries {
		p := filepath.Join(dirs.Versions, e.Name(), "StudioMCP.exe")
		if _, err := os.Stat(p); err == nil {
			mcpPath = p
		}
	}
	if mcpPath == "" {
		slog.Error("StudioMCP.exe not found, please install Studio first")
		return 1
	}

	cmd := pfx.Wine(mcpPath)
	// Use a writer that wraps os.Stdout rather than os.Stdout itself so that
	// wine.Cmd.Start triggers the StdoutPipe copy goroutine (Wine bug 58707).
	cmd.Stdout = io.MultiWriter(os.Stdout)
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		slog.Error("MCP server error", "err", err)
		return 1
	}

	return 0
}
