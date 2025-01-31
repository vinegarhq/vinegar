package main

import (
	"C"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"unsafe"

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

func (ui *ui) ActivateCommandLine(_ gio.Application, cl uintptr) int {
	acl := gio.ApplicationCommandLineNewFromInternalPtr(cl)
	ptr := acl.GetArguments(0)

	// This is the cost of using puregotk and not gotk4.
	var args []string
	for i := 0; ; i++ {
		cStr := (**C.char)(unsafe.Pointer(ptr + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		if *cStr == nil {
			break
		}
		args = append(args, C.GoString(*cStr))
	}

	subcmd := ""
	if len(subcmd) > 2 {
		subcmd = args[1]
	}

	var act func(...string)
	switch subcmd {
	case "run":
		act = ui.ActivateBootstrapper
	case "":
		act = ui.ActivateControl
	default:
		acl.Printerr("Unrecognized subcommand: %s\n", subcmd)
		return 1
	}

	actcb := func(_ gio.Application) {
		act(args[1:]...)
	}

	ui.app.ConnectActivate(&actcb)
	ui.app.Activate()
	return 0
}

func (ui *ui) ActivateControl(_ ...string) {
	err := ui.LoadConfig()
	if err != nil {
		ui.presentSimpleError(err)
		slog.Warn("Falling back to default configuration!")
	}
	ui.NewControl()
}

func (ui *ui) ActivateBootstrapper(args ...string) {
	err := ui.LoadConfig()
	if err != nil {
		ui.presentSimpleError(err)
		return
	}

	b := ui.NewBootstrapper()
	b.win.Show()
	Background(func() {
		go func() {
			b.app.Hold()
			defer b.app.Release()
			err := b.RunArgs(args...)
			if err != nil {
				Background(func() {
					b.presentSimpleError(err)
				})
			}
		}()
	})
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

	ui := ui{
		app: adw.NewApplication(
			"org.vinegarhq.vinegar.Vinegar",
			gio.GApplicationHandlesCommandLineValue,
		),
		state:   &s,
		logFile: lf,
		cfg:     config.Default(),
	}

	clcb := ui.ActivateCommandLine
	ui.app.ConnectCommandLine(&clcb)

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

	var d adw.MessageDialog
	builder.GetObject("error-dialog").Cast(&d)
	d.SetTransientFor(ui.app.GetActiveWindow())
	d.SetApplication(&ui.app.Application)
	defer d.Unref()

	var ccb gio.AsyncReadyCallback
	ccb = func(_ uintptr, res uintptr, _ uintptr) {
		ar := AsyncResultFromInternalPtr(res)
		r := d.ChooseFinish(ar)
		if r == "open" {
			gtk.ShowUri(&d.Window, "file://"+ui.logFile.Name(), 0)
		}
	}
	c := gio.NewCancellable()
	defer c.Unref()

	d.FormatBodyMarkup("<tt>%s</tt>", e.Error())
	d.Choose(c, &ccb, uintptr(unsafe.Pointer(nil)))
}

func (ui *ui) CacheClear() error {
	slog.Info("Removing Cache directory!")
	return os.RemoveAll(dirs.Cache)
}
