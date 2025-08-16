package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/sewnie/rbxweb"
	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/state"
)

type app struct {
	*adw.Application

	cfg   *config.Config
	state *state.State
	pfx   *wine.Prefix
	rbx   *rbxweb.Client
}

func (s *app) unref() {
	s.Unref()
	slog.Info("Goodbye!")
}

func (ui *app) activateCommandLine(_ gio.Application, cl uintptr) int {
	acl := gio.ApplicationCommandLineNewFromInternalPtr(cl)
	args := acl.GetArguments(0)

	subcmd := ""
	if len(args) >= 2 {
		subcmd = args[1]
	}

	switch subcmd {
	case "run":
		ui.activateBootstrapper(args[2:]...)
	case "":
		ui.activateControl()
	default:
		acl.Printerr("Unrecognized subcommand: %s\n", subcmd)
		return 1
	}
	return 0
}

func (ui *app) activateControl() {
	err := ui.loadConfig()
	ui.newControl()
	if err != nil {
		slog.Warn("Falling back to default configuration!")
		ui.error(err)
	}
}

func (ui *app) activateBootstrapper(args ...string) {
	err := ui.loadConfig()
	if err != nil {
		ui.error(err)
		return
	}

	b := ui.newBootstrapper()

	var tf glib.ThreadFunc = func(uintptr) uintptr {
		defer idle(b.win.Destroy)
		if err := b.run(args...); err != nil {
			idle(func() {
				b.error(err)
			})
		}
		return 0
	}
	glib.NewThread("bootstrapper", &tf, 0)
}

func (s *app) loadConfig() error {
	// will fallback to default configuration if there is an error
	cfg, err := config.Load()

	s.pfx = wine.New(
		filepath.Join(dirs.Prefixes, "studio"),
		cfg.Studio.WineRoot,
	)

	s.cfg = cfg

	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return nil
}

func (ui *app) error(e error) {
	slog.Error("Error!", "err", e.Error())

	// In a bootstrapper context, the window is destroyed to show the
	// error instead, which will make the GtkApplication exit.
	ui.Hold()
	d := adw.NewAlertDialog("Something went wrong", e.Error())
	d.AddResponses("okay", "Ok", "open", "Open Log")
	d.SetCloseResponse("okay")
	d.SetDefaultResponse("okay")
	d.SetResponseAppearance("open", adw.ResponseSuggestedValue)

	var ccb gio.AsyncReadyCallback = func(_ uintptr, res uintptr, _ uintptr) {
		defer ui.Release()
		ar := asyncResultFromInternalPtr(res)
		r := d.ChooseFinish(ar)
		slog.Default()
		uri := "file://" + logFile()
		if r == "open" {
			gtk.ShowUri(nil, uri, 0)
		}
	}

	d.Choose(nil, nil, &ccb, 0)
}
