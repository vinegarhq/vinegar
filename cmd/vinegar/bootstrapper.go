package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/apprehensions/rbxbin"
	"github.com/apprehensions/rbxweb/clientsettings"
	"github.com/godbus/dbus/v5"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	cp "github.com/otiai10/copy"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/studiorpc"
)

var Studio = clientsettings.WindowsStudio64

const KillWait = 3 * time.Second

const (
	// Randomly chosen log entry in cases where Studios process
	// continues to run. Due to a lack of bug reports, it is unknown
	// specifically which entry to use for these types of cases.
	StudioShutdownEntry = "[FLog::LifecycleManager] Exited ApplicationScope"
)

type bootstrapper struct {
	*ui
	builder *gtk.Builder
	win     *adw.Window

	pbar   gtk.ProgressBar
	status gtk.Label

	dir string
	bin rbxbin.Deployment

	rp *studiorpc.StudioRPC
}

func (s *ui) NewBootstrapper() *bootstrapper {
	b := bootstrapper{
		builder: gtk.NewBuilderFromResource("/org/vinegarhq/Vinegar/ui/bootstrapper.ui"),
		ui:      s,
		rp:      studiorpc.New(),
	}

	var win adw.Window
	b.builder.GetObject("window").Cast(&win)
	b.win = &win
	b.win.SetApplication(&s.app.Application)
	s.app.AddWindow(&b.win.Window)

	b.builder.GetObject("status").Cast(&b.status)
	b.builder.GetObject("progress").Cast(&b.pbar)
	b.status.Unref()
	b.pbar.Unref()

	b.win.Present()
	b.win.Hide()
	// b.win.Unref()

	return &b
}

func (b *bootstrapper) Performing() func() {
	var tcb glib.SourceFunc
	tcb = func(uintptr) bool {
		b.pbar.Pulse()
		return true
	}
	id := glib.TimeoutAdd(128, &tcb, uintptr(unsafe.Pointer(nil)))
	return func() {
		glib.SourceRemove(id)
	}
}

func (b *bootstrapper) Message(msg string, args ...any) {
	b.status.SetLabel(msg + "...")
	slog.Info(msg, args...)
}

func (b *bootstrapper) Run() error {
	return b.RunArgs()
}

func (b *bootstrapper) RunArgs(args ...string) error {
	if len(args) == 1 && args[0] == "roblox-" {
		b.HandleProtocolURI(args[0])
	}

	if err := b.Setup(); err != nil {
		return fmt.Errorf("setup: %w", err)
	}

	if err := b.Execute(args...); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}

func (b *bootstrapper) Setup() error {
	b.pbar.SetFraction(0.0)
	b.removePlayer()

	if err := b.SetupPrefix(); err != nil {
		return fmt.Errorf("prefix: %w", err)
	}

	if err := b.SetupDeployment(); err != nil {
		return err
	}

	stop := b.Performing()

	b.Message("Applying Environment")
	b.cfg.Studio.Env.Setenv()

	if err := b.SetupOverlay(); err != nil {
		return fmt.Errorf("setup overlay: %w", err)
	}

	b.Message("Applying FFlags")
	if err := b.cfg.Studio.FFlags.Apply(b.dir); err != nil {
		return fmt.Errorf("apply fflags: %w", err)
	}

	stop()

	if err := b.SetupDxvk(); err != nil {
		return fmt.Errorf("setup dxvk %s: %w", b.cfg.Studio.DxvkVersion, err)
	}

	b.Message("Updating State")
	if err := b.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	slog.Info("Successfuly installed!", "guid", b.bin.GUID)

	return nil
}

func (b *bootstrapper) SetupOverlay() error {
	dir := filepath.Join(dirs.Overlays, strings.ToLower(Studio.Short()))

	// Don't copy Overlay if it doesn't exist
	_, err := os.Stat(dir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	b.Message("Copying Overlay")

	return cp.Copy(dir, b.dir)
}

func (b *bootstrapper) removePlayer() {
	// Player is no longer supported by Vinegar, remove unnecessary data
	if b.state.Player.Version != "" || b.state.Player.DxvkVersion != "" {
		os.RemoveAll(filepath.Join(dirs.Versions, b.state.Player.Version))
		os.RemoveAll(filepath.Join(dirs.Prefixes, "player"))
		b.state.Player.DxvkVersion = ""
		b.state.Player.Version = ""
		b.state.Player.Packages = nil
	}
}

func (b *bootstrapper) HandleRobloxLog(line string) {
	if !b.cfg.Studio.Quiet {
		slog.Log(context.Background(), logging.LevelRoblox, line)
	}

	if strings.Contains(line, StudioShutdownEntry) {
		go func() {
			time.Sleep(KillWait)
			syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()
	}

	if b.cfg.Studio.DiscordRPC {
		if err := b.rp.Handle(line); err != nil {
			slog.Error("Presence handling failed", "error", err)
		}
	}
}

func (b *bootstrapper) HandleProtocolURI(mime string) {
	uris := strings.Split(mime, "+")
	for _, uri := range uris {
		kv := strings.Split(uri, ":")

		if len(kv) == 2 && kv[0] == "channel" {
			c := kv[1]
			if c == "" {
				continue
			}

			slog.Warn("Roblox has requested a user channel, changing...", "channel", c)
			b.cfg.Studio.Channel = c
		}
	}
}

func (b *bootstrapper) RegisterGameMode(pid int32) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		slog.Error("Failed to connect to D-Bus", "error", err)
		return
	}

	desktop := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")

	call := desktop.Call("org.freedesktop.portal.GameMode.RegisterGame", 0, pid)
	if call.Err != nil && !errors.Is(call.Err, dbus.ErrMsgNoObject) {
		slog.Error("Failed to register to GameMode", "error", err)
		return
	}
}

func (b *bootstrapper) SetMime() error {
	o, err := exec.Command("xdg-mime", "default",
		"org.vinegarhq.Vinegar.studio.desktop",
		"x-scheme-handler/roblox-studio",
		"x-scheme-handler/roblox-studio-auth",
		"application/x-roblox-rbxl",
		"application/x-roblox-rbxlx",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setup mime: %s", string(o))
	}
	return nil
}
