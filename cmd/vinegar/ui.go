package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/apprehensions/wine"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	slogmulti "github.com/samber/slog-multi"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/state"
)

const errorFormat = "Vinegar has encountered an error: <tt>%v</tt>\nThe log file is shown below for debugging."

type ui struct {
	app *adw.Application

	cfg   *config.Config
	state *state.State
	pfx   *wine.Prefix

	logFile *os.File
}

func Background(bg func()) {
	var idlecb glib.SourceFunc
	idlecb = func(uintptr) bool {
		defer glib.UnrefCallback(&idlecb)
		bg()
		return false
	}
	glib.IdleAdd(&idlecb, 0)
}

func New() ui {
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

	slog.Info("Initializing UI...")

	ui := ui{
		app: adw.NewApplication(
			"org.vinegarhq.vinegar.Vinegar",
			gio.GApplicationFlagsNoneValue,
		),
		state:   &s,
		logFile: lf,
		cfg:     config.Default(),
	}

	ol := gio.NewSimpleAction("logfile-open", nil)
	olcb := func(_ gio.SimpleAction, p uintptr) {
		gtk.ShowUri(ui.app.GetActiveWindow(), "file://"+lf.Name(), 0)
	}
	ol.ConnectActivate(&olcb)
	ui.app.AddAction(ol)
	ol.Unref()

	actcb := func(_ gio.Application) {
		err := ui.LoadConfig()
		if err != nil {
			ui.presentSimpleError(err)
			slog.Warn("Falling back to default configuration!")
		}

		switch flag.Arg(0) {
		case "run":
			if err != nil {
				return
			}
			b := ui.NewBootstrapper()
			Background(func() {
				go func() {
					if err := b.RunArgs(flag.Args()...); err != nil {
						b.presentSimpleError(err)
					}
				}()
			})
		default:
			ui.NewControl()
		}
	}
	ui.app.ConnectActivate(&actcb)

	return ui
}

func (s *ui) LoadConfig() error {
	// will fallback to default configuration if there is an error
	cfg, err := config.Load()

	s.pfx = wine.New(
		filepath.Join(dirs.Prefixes, "studio"),
		s.cfg.Studio.WineRoot,
	)

	s.cfg = cfg

	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return nil
}

func (s *ui) Unref() {
	s.app.Unref()
	s.logFile.Close()
	slog.Info("Goodbye!")
}

func scrollToBottom(tv *gtk.TextView) {
	Background(func() {
		var iter gtk.TextIter
		buf := tv.GetBuffer()
		buf.GetEndIter(&iter)
		buf.Unref()
		tv.ScrollToIter(&iter, 0, true, 0, 1)
	})
}

func (ui *ui) setLogContent(tv *gtk.TextView) {
	_, _ = ui.logFile.Seek(0, 0)
	b, err := io.ReadAll(ui.logFile)
	if err != nil {
		b = []byte(fmt.Sprintf("Failed to read log file for viewing: %v", err))
	}

	buf := tv.GetBuffer()
	buf.SetText(string(b), -1)
	scrollToBottom(tv)
	buf.Unref()
}

func (ui *ui) presentError(e error) {
	builder := gtk.NewBuilderFromString(resource("error.ui"), -1)
	defer builder.Unref()

	slog.Error(e.Error())

	var win adw.Window
	builder.GetObject("window").Cast(&win)
	ui.app.AddWindow(&win.Window)

	var label gtk.Label
	builder.GetObject("error-label").Cast(&label)
	label.SetMarkup(fmt.Sprintf(errorFormat, e))
	label.Unref()

	var tv gtk.TextView
	builder.GetObject("log-output").Cast(&tv)
	ui.setLogContent(&tv)
	tv.Unref()

	win.SetTitle("Error Report")
	win.SetDefaultSize(512, 320)
	win.Present()
	win.Unref()
}

func (ui *ui) presentSimpleError(e error) {
	builder := gtk.NewBuilderFromString(resource("error.ui"), -1)
	defer builder.Unref()

	var w adw.Window
	builder.GetObject("dialog").Cast(&w)
	w.SetTransientFor(ui.app.GetActiveWindow())
	w.SetApplication(&ui.app.Application)

	var msg gtk.Label
	builder.GetObject("error").Cast(&msg)
	msg.SetLabel(e.Error())
	msg.Unref()

	w.Present()
	w.Unref()
}
