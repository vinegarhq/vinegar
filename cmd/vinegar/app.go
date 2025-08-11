package main

import (
	"fmt"
	"log/slog"
	"os"
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

	logFile *os.File
}

func (s *app) unref() {
	s.Unref()
	s.logFile.Close()
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
	if err != nil {
		ui.error(err)
		slog.Warn("Falling back to default configuration!")
	}
	ui.newControl()
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
			idle(func() { b.error(err) })
		}
		return null
	}
	glib.NewThread("bootstrapper", &tf, null)
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
	builder := gtk.NewBuilderFromResource("/org/vinegarhq/Vinegar/ui/error.ui")
	defer builder.Unref()

	var d adw.MessageDialog
	builder.GetObject("error-dialog").Cast(&d)
	// It is unreccomended to have a AdwMessageDialog without a
	// parent, and opening the log file without the parent
	// will be impossible, this is fine, since the error in
	// such contexts does not need further information.
	win := ui.GetActiveWindow()
	if win != nil {
		d.SetTransientFor(ui.GetActiveWindow())
	}
	d.SetApplication(&ui.Application.Application)
	defer d.Unref()

	slog.Error("Error!", "err", e)

	if win == nil {
		d.AddResponses("okay", "Ok")
	} else {
		d.AddResponses("okay", "Ok", "open", "Open Log")
	}

	var ccb gio.AsyncReadyCallback = func(_ uintptr, res uintptr, _ uintptr) {
		ar := asyncResultFromInternalPtr(res)
		r := d.ChooseFinish(ar)
		if win != nil && r == "open" {
			gtk.ShowUri(&d.Window, "file://"+ui.logFile.Name(), 0)
		}
	}

	c := gio.NewCancellable()
	defer c.Unref()

	d.FormatBodyMarkup("%s", e.Error())
	d.Choose(c, &ccb, null)
}
