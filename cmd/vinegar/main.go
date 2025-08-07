package main

import (
	"log"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	slogmulti "github.com/samber/slog-multi"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/state"
)

var version string

func main() {
	debug.SetPanicOnFault(true)

	s, err := state.Load()
	if err != nil {
		log.Fatalf("load state: %s", err)
	}

	lf, err := logging.NewFile()
	if err != nil {
		log.Fatalf("log file: %s", err)
	}

	slog.SetDefault(slog.New(slogmulti.Fanout(
		logging.NewTextHandler(os.Stderr, false),
		logging.NewTextHandler(lf, true),
	)))

	ui := app{
		Application: adw.NewApplication(
			"org.vinegarhq.Vinegar",
			gio.GApplicationHandlesCommandLineValue,
		),
		state:   &s,
		logFile: lf,
		cfg:     config.Default(),
	}
	defer ui.unref()

	aclcb := ui.activateCommandLine
	ui.ConnectCommandLine(&aclcb)

	if code := ui.Run(len(os.Args), os.Args); code > 0 {
		os.Exit(code)
	}
}
