package main

import (
	"log/slog"
	"os"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/pojntfx/go-gettext/pkg/i18n"
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

	// Detect MCP mode before GApplication so that GApplicationNonUniqueValue
	// can be set, preventing GLib from forwarding the command line to an
	// existing primary instance (which would have the wrong stdio for MCP).
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "run" {
		args = args[1:]
	}
	mcpMode := len(args) == 1 && args[0] == "mcp"

	if code := newApp(mcpMode).Run(int32(len(os.Args)), os.Args); code > 0 {
		os.Exit(int(code))
	}
}
