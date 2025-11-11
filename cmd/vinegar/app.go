package main

import (
	"context"
	"errors"
	"fmt"
	"io"
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

	mgr  *manager // nullable
	boot *bootstrapper

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
	pfx, err := a.cfg.Prefix()
	if err != nil {
		return fmt.Errorf("prefix configure: %w", err)
	}
	pfx.Stderr = io.Writer(a)
	pfx.Stdout = pfx.Stderr

	if a.cfg.Debug {
		a.rbx.Client.Transport = &debugTransport{
			underlying: http.DefaultTransport,
		}
	}

	a.pfx = pfx
	return nil
}

func (a *app) startup(_ gio.Application) {
	slog.SetDefault(slog.New(
		logging.NewHandler(os.Stderr, slog.LevelInfo)))

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
		return
	}
	a.cfg = cfg

	if err := a.reload(); err != nil {
		a.showError(err)
	}
}

func (a *app) commandLine(_ gio.Application, clPtr uintptr) int {
	if a.cfg == nil || a.pfx == nil {
		return 1
	}

	cl := gio.ApplicationCommandLineNewFromInternalPtr(clPtr)
	args := cl.GetArguments(0)[1:] // skip argv0
	if len(args) >= 1 && args[0] == "run" {
		args = args[1:] // skip 'run' cmd
	}

	if len(args) == 1 && args[0] == "manage" {
		if a.mgr == nil {
			a.mgr = a.newManager()
		}
		a.mgr.win.Present()
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
		return fmt.Errorf("browser login set: %w", err)
	}
	return nil
}

func (a *app) Write(b []byte) (int, error) {
	for line := range strings.SplitSeq(string(b[:len(b)-1]), "\n") {
		// XXXX:channel:class OutputDebugStringA "[FLog::Foo] Message"
		if a.boot != nil && len(line) >= 39 && line[19:37] == "OutputDebugStringA" {
			// Avoid "\n" calls to OutputDebugStringA
			if len(line) >= 44 {
				a.boot.handleRobloxLog(line[39 : len(line)-1])
			}
			return len(b), nil
		}

		a.handleWineLog(line)
	}
	return len(b), nil
}

func (a *app) handleWineLog(line string) {
	if strings.Contains(line, "to unimplemented function advapi32.dll.SystemFunction036") {
		err := errors.New("Your Wineprefix is corrupt! Please delete all data in Vinegar's settings.")
		gtkutil.IdleAdd(func() {
			a.pfx.Server(wine.ServerKill, "9")
			a.showError(err)
		})
	}

	slog.Log(context.Background(), logging.LevelWine.Level(), line)
}

// GetPrefixForDesktop returns the appropriate prefix for a given desktop
func (a *app) GetPrefixForDesktop(desktopName string) *wine.Prefix {
	if a.cfg.DesktopManager.Enabled && desktopName != "" {
		if pfx, exists := a.prefixes[desktopName]; exists {
			return pfx
		}
		// If desktop doesn't exist, create it on-demand
		if pfx, err := a.cfg.PrefixForDesktop(desktopName); err == nil {
			pfx.Stderr = io.Writer(a)
			pfx.Stdout = pfx.Stderr
			a.prefixes[desktopName] = pfx
			return pfx
		}
	}
	// Fallback to default prefix
	return a.pfx
}

func (a *app) AssignDesktopForInstance() string {
	if !a.cfg.DesktopManager.Enabled {
		return "studio"
	}
	return a.cfg.AssignDesktop()
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

	var ccb gio.AsyncReadyCallback = func(_ uintptr, resPtr uintptr, _ uintptr) {
		defer a.Release()
		res := gio.SimpleAsyncResultNewFromInternalPtr(resPtr)
		r := d.ChooseFinish(res)
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
