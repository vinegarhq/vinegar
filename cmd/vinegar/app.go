package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"codeberg.org/puregotk/puregotk/v4/adw"
	"codeberg.org/puregotk/puregotk/v4/gio"
	"codeberg.org/puregotk/puregotk/v4/glib"
	"codeberg.org/puregotk/puregotk/v4/gtk"
	"github.com/sewnie/rbxweb"
	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/sysinfo"
	"golang.org/x/sys/unix"

	. "github.com/pojntfx/go-gettext/pkg/i18n"
)

type app struct {
	*adw.Application

	// Current latest version obtained from metainfo.xml
	version string
	mcpMode bool

	cfg *config.Config
	pfx *wine.Prefix
	rbx *rbxweb.Client
	bus *gio.DBusConnection // nullable

	mgr  *manager // nullable
	boot *bootstrapper
}

func newApp(mcpMode bool) *app {
	b := gutil.ResourceData(gutil.Resource("metainfo.xml"))
	data := struct {
		XMLName  xml.Name `xml:"component"`
		Releases struct {
			Release []struct {
				Text    string `xml:",chardata"`
				Version string `xml:"version,attr"`
				Date    string `xml:"date,attr"`
			} `xml:"release"`
		} `xml:"releases"`
	}{}

	if err := xml.Unmarshal(b, &data); err != nil {
		log.Panicln("expected valid appstream:", err)
	}

	// command-line is preferred over open due to open abstracting the real
	// argument to GFile, which is not an effective wrapper for Studio arguments.
	// NonUnique is added in MCP mode so this process is never forwarded to an
	// existing primary instance (which has the wrong stdio for binary MCP piping).
	flags := gio.GApplicationHandlesCommandLineValue
	if mcpMode {
		flags |= gio.GApplicationNonUniqueValue
	}

	a := app{
		Application: adw.NewApplication("org.vinegarhq.Vinegar", flags),
		version:     data.Releases.Release[0].Version,
		rbx:         rbxweb.NewClient(),
		mcpMode:     mcpMode,
	}

	startup := a.startup
	a.ConnectStartup(&startup)
	cli := a.commandLine
	a.ConnectCommandLine(&cli)
	shutdown := a.shutdown
	a.ConnectShutdown(&shutdown)

	return &a
}

func (a *app) applyConfig() {
	// https://github.com/vinegarhq/vinegar/issues/746
	// Check currently initialized prefix against the new configuration
	if a.pfx != nil && a.pfx.Root != string(a.cfg.Studio.WineRoot) && a.pfx.Running() {
		slog.Info("Wine installation changed, killing Wine", "err", a.pfx.Kill())
	}
	a.pfx = a.cfg.Prefix()
	a.pfx.Stderr = io.Writer(a)
	a.pfx.Stdout = a.pfx.Stderr

	if a.cfg.Debug {
		a.rbx.Client.Transport = &debugTransport{
			underlying: http.DefaultTransport,
		}
	}

	if string(a.cfg.Studio.WineRoot) == dirs.WinePath {
		return
	}

	// User is currently using another wine build
	path, err := filepath.EvalSymlinks(dirs.WinePath)
	if err == nil {
		slog.Info("Removing unused Wine build", "path", path)
		_ = os.RemoveAll(path)
		_ = os.RemoveAll(dirs.WinePath)
	}
}

func (a *app) startup(_ gio.Application) {
	slog.SetDefault(slog.New(
		logging.NewHandler(os.Stderr, slog.LevelInfo)))

	if !a.mcpMode {
		slog.Info("System information",
			"cpu", sysinfo.CPU.Name,
			"mem", glib.FormatSizeForDisplay(int64(sysinfo.Memory)),
			"distro", sysinfo.Distro,
			"display", sysinfo.Display)
		slog.Info("DRM Devices",
			"cards", sysinfo.Cards)

		// ChromeOS allocates 4GB [citation needed] for Crostini.
		if sysinfo.Memory < 4*1024*1024 && !a.cfg.Debug {
			a.showError(errors.New(L(
				"This system does not meet the minimum requirements to run Roblox Studio." +
					"It is recommended to run Vinegar with sufficient memory and graphics.")))
			return
		}

		a.boot = a.newBootstrapper()

		// Required for GameMode
		conn, err := gio.BusGetSync(gio.GBusTypeSessionValue, nil)
		if err != nil {
			slog.Error("Failed to retrieve session bus, all DBus operations will be ignored", "err", err)
		} else {
			a.bus = conn
		}
	}

	// Any error that can occur here is I/O failure, which
	// is unrecoverable, unlike the user editing the configuration
	// manually, which is user error, as the manager validates the
	// configuration during editing.
	var err error
	a.cfg, err = config.Load()
	if err != nil {
		if a.mcpMode {
			slog.Error("Failed to load config", "err", err)
			a.Quit()
			return
		}
		a.showError(fmt.Errorf("config error: %w", err))
		return
	}
	a.applyConfig()

	if !a.mcpMode {
		sm := a.GetStyleManager()
		cb := a.updateWineTheme
		sm.ConnectSignal("notify::dark", &cb)
	}
}

func (a *app) commandLine(_ gio.Application, clPtr uintptr) int32 {
	if a.cfg == nil || a.pfx == nil {
		return 1
	}

	cl := gio.ApplicationCommandLineNewFromInternalPtr(clPtr)
	args := cl.GetArguments(nil)[1:] // skip argv0
	if len(args) >= 1 && args[0] == "run" {
		args = args[1:] // skip 'run' cmd
	}

	if len(args) == 1 && args[0] == "mcp" {
		a.Hold()
		go func() {
			if err := a.runMCP(); err != nil {
				slog.Error("MCP server error", "err", err)
			}
			a.Release()
			gutil.IdleAdd(func() { a.Quit() })
		}()
		return 0
	}

	// Override arguments to prioritize welcome screen
	_, err := os.Stat(dirs.Data)
	if err != nil {
		err := os.MkdirAll(dirs.Data, 0o755)
		if err != nil {
			slog.Error("Failed to initialize data directory", "err", err)
			return 1
		}
	}

	if err != nil || (len(args) == 1 &&
		(args[0] == "manage" || args[0] == "config")) {
		// Prevent multiple windows of manager
		// existing at once
		if a.mgr == nil {
			a.mgr = a.newManager()
		}
		a.mgr.win.Present()

		if err != nil {
			view := gutil.
				GetObject[adw.NavigationView](a.mgr.builder, "navigation")
			view.PushByTag("welcome")

		}
		return 0
	}

	// As bootstrapper runs in a thread, it may add itself as a
	// window too late for Gtk to register. Prevent GtkApplication
	// from closing for when there are no windows.
	a.Hold()
	a.errThread(func() error {
		defer a.Release()
		return a.boot.run(args[:]...)
	})
	return 0
}

func (a *app) shutdown(_ gio.Application) {
	if a.boot != nil {
		if err := a.boot.backupSettings(); err != nil {
			slog.Error("Failed to backup Studio settings", "err", err)
		}

		// TODO: Wait for studio processes or Wineserver to die
		if a.boot.count > 1 {
			slog.Warn("Handing off Wineserver control!")
			return
		}
	}
	slog.Info("Goodbye!")
}

func (a *app) appInfo() *gio.AppInfoBase {
	for app := range gutil.List[gio.AppInfoBase](gio.AppInfoGetAll()) {
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

func (a *app) errThread(fn func() error) {
	go func() {
		if err := fn(); err != nil {
			gutil.IdleAdd(func() {
				a.showError(err)
			})
		}
	}()
}

func (a *app) showError(e error) {
	slog.Error("Error!", "err", e.Error())

	// In a bootstrapper context, the window is destroyed to show the
	// error instead, which will make the GtkApplication exit.
	a.Hold()
	d := adw.NewAlertDialog(L("Something went wrong"), e.Error())
	d.AddResponses("okay", L("Ok"), "open", L("Open Log"))
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

func (a *app) registerGame(processHandle uintptr) error {
	if a.bus == nil {
		return nil
	}

	req, err := unix.PidfdOpen(os.Getpid(), 0)
	if err != nil {
		return err
	}

	resp, err := a.bus.CallWithUnixFdListSync("org.freedesktop.portal.Desktop",
		"/org/freedesktop/portal/desktop",
		"org.freedesktop.portal.GameMode",
		"RegisterGameByPIDFd",
		glib.NewVariant("(hh)", processHandle, req),
		glib.NewVariantType("(i)"),
		gio.GDbusCallFlagsNoneValue,
		-1,
		gio.NewUnixFDListFromArray([]int32{int32(processHandle), int32(req)}, 2),
		nil,
		nil,
	)
	_ = unix.Close(req)
	if err != nil {
		return fmt.Errorf("dbus: %w", err)
	}

	var res int32
	resp.Get("(i)", &res)
	if res < 0 {
		// The Gamemode is proxied through the XDG portal, so if the gamemode
		// daemon is not running, the only response returned from the XDG portal
		// is that the process has been rejected.
		return errors.New("rejected")
	}
	slog.Info("Registered with GameMode", "response", res, "pidfd", processHandle)

	return nil
}

func (a *app) runMCP() error {
	if !a.pfx.Exists() {
		return errors.New("Wine prefix does not exist, please install Studio first")
	}
	if !a.pfx.Running() {
		return errors.New("Studio is not running")
	}

	entries, err := os.ReadDir(dirs.Versions)
	if err != nil {
		return fmt.Errorf("read versions: %w", err)
	}

	var mcpPath string
	for _, e := range entries {
		p := filepath.Join(dirs.Versions, e.Name(), "StudioMCP.exe")
		if _, err := os.Stat(p); err == nil {
			mcpPath = p
		}
	}
	if mcpPath == "" {
		return errors.New("StudioMCP.exe not found, please install Studio first")
	}

	cmd := a.pfx.Wine(mcpPath)
	// Use a writer that wraps os.Stdout rather than os.Stdout itself so that
	// wine.Cmd.Start triggers the StdoutPipe copy goroutine (Wine bug 58707).
	cmd.Stdout = io.MultiWriter(os.Stdout)
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	// GLib only auto-handles SIGTERM, not SIGINT. Intercept both signals and
	// kill the Wine process so cmd.Wait() unblocks and GApplication can exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(sig)
		close(sig)
	}()
	go func() {
		if _, ok := <-sig; ok {
			_ = cmd.Process.Kill()
		}
	}()

	return cmd.Wait()
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
