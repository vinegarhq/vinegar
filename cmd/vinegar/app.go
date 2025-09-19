package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/sewnie/rbxweb"
	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/state"
)

type app struct {
	*adw.Application

	cfg   *config.Config
	state *state.State
	pfx   *wine.Prefix
	rbx   *rbxweb.Client

	// initialized only in Application::command-line
	ctl  *control
	boot *bootstrapper // also set if control runs boot

	keepLog bool
}

func newApp() *app {
	s, err := state.Load()
	if err != nil {
		log.Fatalf("load state: %s", err)
	}

	a := app{
		Application: adw.NewApplication(
			"org.vinegarhq.Vinegar",
			// command-line is preferred over open due to open
			// abstracting real argument to GFile, which is not
			// an effective wrapper for Studio arguments.
			gio.GApplicationHandlesCommandLineValue,
		),
		state: &s,
		cfg:   config.Default(),
		rbx:   rbxweb.NewClient(),
	}

	clcb := a.commandLine
	a.ConnectCommandLine(&clcb)
	scb := a.shutdown
	a.ConnectShutdown(&scb)

	dialogA := gio.NewSimpleAction("show-login-dialog", nil)
	dialobCb := func(_ gio.SimpleAction, _ uintptr) {
		a.newLogin()
	}
	dialogA.ConnectActivate(&dialobCb)
	a.AddAction(dialogA)

	return &a
}

func (a *app) shutdown(_ gio.Application) {
	if !a.keepLog && !a.cfg.Debug {
		_ = os.Remove(logging.Path)
	}
	slog.Info("Goodbye!")
}

func (a *app) commandLine(_ gio.Application, clPtr uintptr) int {
	cl := gio.ApplicationCommandLineNewFromInternalPtr(clPtr)
	args := cl.GetArguments(0)

	err := a.loadConfig()

	if len(args) < 2 {
		if err != nil {
			a.error(err)
		}
		if a.ctl == nil {
			a.ctl = a.newControl()
		}
		a.ctl.win.Present()
		return 0
	}

	if err != nil {
		a.error(err)
		return 22
	}

	if args[1] == "run" { // backwards compatibility
		args = args[1:]
	}

	if a.boot == nil {
		a.boot = a.newBootstrapper()
	}
	var tf glib.ThreadFunc = func(uintptr) uintptr {
		if err := a.boot.run(args[1:]...); err != nil {
			idle(func() {
				a.boot.error(err)
			})
		}
		return 0
	}
	glib.NewThread("bootstrapper", &tf, 0)

	return 0
}

// Write implements io.Writer for app and is used to exclusively send all
// data recieved to the log under the WINE log level.
func (a *app) Write(b []byte) (int, error) {
	for line := range strings.SplitSeq(string(b), "\n") {
		if line == "" {
			continue
		}

		// XXXX:channel:class OutputDebugStringA "[FLog::Foo] Message"
		if a.boot != nil && len(line) >= 39 && line[19:37] == "OutputDebugStringA" {
			// Avoid "\n" calls to OutputDebugStringA
			if len(line) >= 87 {
				a.boot.handleRobloxLog(line[39 : len(line)-1])
			}
			continue
		}

		slog.Log(context.Background(), logging.LevelWine, line)
	}
	return len(b), nil
}

func (a *app) loadConfig() error {
	// will fallback to default configuration if there is an error
	cfg, err := config.Load()

	a.pfx = wine.New(
		filepath.Join(dirs.Prefixes, "studio"),
		cfg.Studio.WineRoot,
	)
	a.pfx.Stderr = a
	a.pfx.Stdout = a

	a.cfg = cfg

	if cfg.Debug {
		a.rbx.Client.Transport = &debugTransport{
			underlying: http.DefaultTransport,
		}
	}

	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return nil
}

func (a *app) error(e error) {
	a.keepLog = true
	slog.Error("Error!", "err", e.Error())

	// In a bootstrapper context, the window is destroyed to show the
	// error instead, which will make the GtkApplication exit.
	a.Hold()
	d := adw.NewAlertDialog("Something went wrong", e.Error())
	d.AddResponses("okay", "Ok", "open", "Open Log")
	d.SetCloseResponse("okay")
	d.SetDefaultResponse("okay")
	d.SetResponseAppearance("open", adw.ResponseSuggestedValue)

	var ccb gio.AsyncReadyCallback = func(_ uintptr, res uintptr, _ uintptr) {
		defer a.Release()
		ar := asyncResultFromInternalPtr(res)
		r := d.ChooseFinish(ar)
		slog.Default()
		uri := "file://" + logging.Path
		if r == "open" {
			gtk.ShowUri(nil, uri, 0)
		}
	}

	d.Choose(nil, nil, &ccb, 0)
}

type debugTransport struct {
	underlying http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	slog.Debug("rbxweb request",
		"method", req.Method,
		"url", req.URL.String(),
	)

	resp, err := t.underlying.RoundTrip(req)
	if err != nil {
		slog.Debug("rbxweb request failed", "error", err)
		return nil, err
	}

	slog.Debug("rbxweb response", "status", resp.StatusCode)
	return resp, nil
}
