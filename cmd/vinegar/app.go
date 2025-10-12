package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/sewnie/rbxweb"
	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/gtkutil"
	"github.com/vinegarhq/vinegar/internal/logging"
)

type app struct {
	*adw.Application

	cfg *config.Config
	pfx *wine.Prefix
	rbx *rbxweb.Client
	bus *gio.DBusConnection // nullable

	// initialized only in Application::command-line
	ctl  *control      // nullable
	boot *bootstrapper // also set if control runs boot

	keepLog bool
}

func newApp() *app {
	a := app{
		Application: adw.NewApplication(
			"org.vinegarhq.Vinegar",
			// command-line is preferred over open due to open
			// abstracting real argument to GFile, which is not
			// an effective wrapper for Studio arguments.
			gio.GApplicationHandlesCommandLineValue,
		),
		cfg: config.Default(),
		rbx: rbxweb.NewClient(),
	}

	startup := a.startup
	a.ConnectStartup(&startup)
	cli := a.commandLine
	a.ConnectCommandLine(&cli)
	shutdown := a.shutdown
	a.ConnectShutdown(&shutdown)

	return &a
}

func (a *app) reload() error {
	slog.Info("Reloading!")

	pfx, err := a.cfg.Prefix()
	if err != nil {
		return fmt.Errorf("prefix configure: %w", err)
	}
	pfx.Stderr = a
	pfx.Stdout = a

	if a.cfg.Debug {
		a.rbx.Client.Transport = &debugTransport{
			underlying: http.DefaultTransport,
		}
	}

	a.pfx = pfx
	return nil
}

func (a *app) startup(_ gio.Application) {
	a.boot = a.newBootstrapper()

	conn, err := gio.BusGetSync(gio.GBusTypeSessionValue, nil)
	if err != nil {
		slog.Error("Failed to retrieve session bus, all DBus operations will be ignored", "err", err)
	} else {
		a.bus = conn
	}

	cfg, err := config.Load()
	if err != nil {
		a.showError(fmt.Errorf("config error: %w", err))
	} else {
		a.cfg = cfg
	}

	if err := a.reload(); err != nil {
		a.showError(err)
	}
}

func (a *app) commandLine(_ gio.Application, clPtr uintptr) int {
	if a.keepLog {
		// Error dialog is open currently
		return 0
	}

	cl := gio.ApplicationCommandLineNewFromInternalPtr(clPtr)
	args := cl.GetArguments(0)[1:] // skip argv0
	if len(args) >= 1 && args[0] == "run" {
		args = args[1:] // skip 'run' cmd
	}

	if len(args) == 1 && args[0] == "config" {
		if a.ctl == nil {
			a.ctl = a.newControl()
		}
		a.ctl.win.Present()
		return 0
	}

	a.Hold()
	a.errThread(func() error {
		defer a.Release()
		return a.boot.run(args[:]...)
	})
	return 0
}

func (a *app) shutdown(_ gio.Application) {
	if !a.keepLog && !a.cfg.Debug {
		_ = os.Remove(logging.Path)
	}
	slog.Info("Goodbye!")
}

func (a *app) appInfo() *gio.AppInfoBase {
	for app := range gtkutil.List[gio.AppInfoBase](gio.AppInfoGetAll()) {
		if strings.HasPrefix(app.GetId(), a.GetApplicationId()) {
			return app
		}
	}
	return nil
}

func (a *app) setMime() error {
	selfApp := a.appInfo()
	if selfApp == nil {
		return errors.New("Where is Vinegar's desktop file? Is this a proper installation?")
	}

	slog.Info("Setting as default application for browser login")
	ok, err := selfApp.SetAsDefaultForType("x-scheme-handler/roblox-studio-auth")
	if !ok || err != nil {
		return fmt.Errorf("Cannot gurantee browser login: %w", err)
	}
	return nil
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

		slog.Log(context.Background(), logging.LevelWine.Level(), line)
	}
	return len(b), nil
}

func (a *app) errThread(fn func() error) {
	go func() {
		if err := fn(); err != nil {
			gtkutil.IdleAdd(func() {
				a.showError(err)
			})
		}
	}()
}

func (a *app) showError(e error) {
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
		ar := gtkutil.AsyncResultFromInternalPtr(res)
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
