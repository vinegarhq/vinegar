package main

import (
	"log"
	"os"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/sewnie/rbxweb"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/state"
	"github.com/vinegarhq/vinegar/sysinfo"
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

	s, err := state.Load()
	if err != nil {
		log.Fatalf("load state: %s", err)
	}

	ui := app{
		Application: adw.NewApplication(
			"org.vinegarhq.Vinegar",
			gio.GApplicationHandlesCommandLineValue,
		),
		state: &s,
		cfg:   config.Default(),
		rbx:   rbxweb.NewClient(),
	}
	defer ui.unref()

	dialogA := gio.NewSimpleAction("show-login-dialog", nil)
	dialobCb := func(a gio.SimpleAction, p uintptr) {
		ui.newLogin()
	}
	dialogA.ConnectActivate(&dialobCb)
	ui.AddAction(dialogA)
	dialogA.Unref()

	aclcb := ui.activateCommandLine
	ui.ConnectCommandLine(&aclcb)

	// TODO: sometimes segfaults for no reason
	if code := ui.Run(len(os.Args), os.Args); code > 0 {
		os.Exit(code)
	}
}
