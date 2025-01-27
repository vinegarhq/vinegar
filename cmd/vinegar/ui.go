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
	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
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

	lf, err := LogFile()
	if err != nil {
		log.Fatalf("log file: %s", err)
	}

	slog.SetDefault(slog.New(slogmulti.Fanout(
		tint.NewHandler(os.Stderr, nil),
		tint.NewHandler(lf, &tint.Options{NoColor: true}),
	)))

	ui := ui{
		app: adw.NewApplication(
			"org.vinegarhq.vinegar.Vinegar",
			gio.GApplicationFlagsNoneValue,
		),
		state:   &s,
		logFile: lf,
	}

	actcb := func(_ gio.Application) {
		if err := ui.LoadConfig(); err != nil {
			ui.presentError(err)
			return
		}

		switch flag.Arg(0) {
		case "run":
			b := ui.NewBootstrapper()
			Background(func() {
				go func() {
					if err := b.RunArgs(flag.Args()...); err != nil {
						b.presentError(err)
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
		cfg.Studio.WineRoot,
	)

	if err != nil {
		return err
	}

	s.cfg = &cfg

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

	var win adw.Window
	builder.GetObject("error").Cast(&win)
	ui.app.AddWindow(&win.Window)

	var label gtk.Label
	builder.GetObject("error_label").Cast(&label)
	label.SetMarkup(fmt.Sprintf(errorFormat, e))
	label.Unref()

	var tv gtk.TextView
	builder.GetObject("log_output").Cast(&tv)
	ui.setLogContent(&tv)
	tv.Unref()

	win.SetTitle("Error Report")
	win.SetDefaultSize(512, 320)
	win.Present()
	win.Unref()
}
