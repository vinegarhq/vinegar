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

var null = uintptr(unsafe.Pointer(nil))

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
	if len(args) >= 2 {
		subcmd = args[1]
	}

	switch subcmd {
	case "run":
		ui.ActivateBootstrapper(args[2:]...)
	case "":
		ui.ActivateControl()
	default:
		acl.Printerr("Unrecognized subcommand: %s\n", subcmd)
		return 1
	}
	return 0
}

func (ui *ui) ActivateControl() {
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
	ui.app.Hold()

	var tf glib.ThreadFunc
	tf = func(uintptr) uintptr {
		defer Background(ui.app.Release)
		if err := b.RunArgs(args...); err != nil {
			Background(func() { b.presentSimpleError(err) })
		}
		return null
	}
	glib.NewThread("bootstrapper", &tf, null)
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

func (ui *ui) presentSimpleError(e error) {
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
		ar := AsyncResultFromInternalPtr(res)
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

func (ui *ui) AboutWindow() *adw.AboutWindow {
	w := adw.NewAboutWindow()
	w.SetApplicationName("Vinegar")
	w.SetApplicationIcon("org.vinegarhq.Vinegar")
	w.SetIssueUrl("https://github.com/vinegarhq/vinegar/issues")
	w.SetSupportUrl("https://discord.gg/dzdzZ6Pps2")
	w.SetWebsite("https://vinegarhq.org")
	w.SetVersion(Version)
	return w
}
