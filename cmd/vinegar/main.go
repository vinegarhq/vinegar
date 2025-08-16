package main

import (
	"log"
	"os"
	"runtime/debug"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/sewnie/rbxweb"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/state"
)

var version string

func main() {
	debug.SetPanicOnFault(true)

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
