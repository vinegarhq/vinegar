package main

import (
	"os"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/vinegarhq/vinegar/internal/sysinfo"
)

var version string

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

	if code := newApp().Run(len(os.Args), os.Args); code > 0 {
		os.Exit(code)
	}
}
