package main

import (
	"C"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/apprehensions/wine"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/state"
)

var null = uintptr(unsafe.Pointer(nil))

const errorFormat = "Vinegar has encountered an error: <tt>%v</tt>\nThe log file is shown below for debugging."

type ui struct {
	app *adw.Application

	cfg   *config.Config
	state *state.State
	pfx   *wine.Prefix

	logFile *os.File
}

func (s *ui) unref() {
	s.app.Unref()
	s.logFile.Close()
	slog.Info("Goodbye!")
}

func idle(bg func()) {
	var idlecb glib.SourceFunc
	idlecb = func(uintptr) bool {
		defer glib.UnrefCallback(&idlecb)
		bg()
		return false
	}
	glib.IdleAdd(&idlecb, 0)
}

func (ui *ui) activateCommandLine(_ gio.Application, cl uintptr) int {
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

func (ui *ui) activateControl() {
	err := ui.loadConfig()
	if err != nil {
		ui.error(err)
		slog.Warn("Falling back to default configuration!")
	}
	ui.newControl()
}

func (ui *ui) activateBootstrapper(args ...string) {
	err := ui.loadConfig()
	if err != nil {
		ui.error(err)
		return
	}

	b := ui.newBootstrapper()
	ui.app.Hold()

	var tf glib.ThreadFunc
	tf = func(uintptr) uintptr {
		defer idle(ui.app.Release)
		if err := b.run(args...); err != nil {
			idle(func() { b.error(err) })
		}
		return null
	}
	glib.NewThread("bootstrapper", &tf, null)
}

func (s *ui) loadConfig() error {
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

func (ui *ui) error(e error) {
	builder := gtk.NewBuilderFromResource("/org/vinegarhq/Vinegar/ui/error.ui")
	defer builder.Unref()

	var d adw.MessageDialog
	builder.GetObject("error-dialog").Cast(&d)
	// It is unreccomended to have a AdwMessageDialog without a
	// parent, and opening the log file without the parent
	// will be impossible, this is fine, since the error in
	// such contexts does not need further information.
	win := ui.app.GetActiveWindow()
	if win != nil {
		d.SetTransientFor(ui.app.GetActiveWindow())
	}
	d.SetApplication(&ui.app.Application)
	defer d.Unref()

	slog.Error("Error!", "err", e)

	if win == nil {
		d.AddResponses("okay", "Ok")
	} else {
		d.AddResponses("okay", "Ok", "open", "Open Log")
	}

	var ccb gio.AsyncReadyCallback
	ccb = func(_ uintptr, res uintptr, _ uintptr) {
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
