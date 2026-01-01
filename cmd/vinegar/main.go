package main

import (
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

	if err := i18n.InitI18n("vinegar", LocaleDir); err != nil {
		panic(err)
	}

	if code := newApp().Run(len(os.Args), os.Args); code > 0 {
		os.Exit(code)
	}
}
