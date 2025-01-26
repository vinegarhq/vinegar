package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/internal/state"
)

const errorFormat = "Vinegar has encountered an error while bootstrapping: <tt>%v</tt>\nThe log file is shown below for debugging."

type ui struct {
	app *adw.Application

	cfg   *config.Config
	state *state.State

	logFile *os.File
}

func New() ui {
	s, err := state.Load()
	if err != nil {
		log.Fatalf("load state: %s", err)
	}

	cfg, err := config.Load(ConfigPath)
	if err != nil {
		log.Fatalf("load config: %s", err)
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
		cfg: &cfg,
		state: &s,
		logFile: lf,
	}

	actcb := func(_ gio.Application) {
		switch flag.Arg(0) {
		case "run":
			b := ui.NewBootstrapper()
			b.Start(flag.Args()...)
		default:
			ui.NewControl()
		}
	}
	ui.app.ConnectActivate(&actcb)

	return ui
}

func (s *ui) Unref() {
	s.app.Unref()
	s.logFile.Close()
	slog.Info("Goodbye!")
}

func scrollToBottom(tv *gtk.TextView) {
	var idlecb glib.SourceFunc
	idlecb = func(uintptr) bool {
		defer glib.UnrefCallback(&idlecb)
		var iter gtk.TextIter
		buf := tv.GetBuffer()
		buf.GetEndIter(&iter)
		buf.Unref()
		tv.ScrollToIter(&iter, 0, true, 0, 1)
		return false
	}
	glib.IdleAdd(&idlecb, 0)
}

func (ui *ui) setLogContent(tv *gtk.TextView) {
	b, err := io.ReadAll(ui.logFile)
	if err != nil {
		b = []byte(fmt.Sprintf("Failed to read log file for viewing: %v", err))
	}

	buf := tv.GetBuffer()
	buf.SetText(string(b), len(string(b)))
	scrollToBottom(tv)
	buf.Unref()
}

func (ui *ui) presentError(e error) {
	builder := gtk.NewBuilderFromString(resource("error.ui"), -1)

	var win adw.Window
	builder.GetObject("error").Cast(&win)
	ui.app.AddWindow(&win.Window)

	destroy := func(_ gtk.Window) bool {
		ui.app.Quit()
		return false
	}
	win.ConnectCloseRequest(&destroy)

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
