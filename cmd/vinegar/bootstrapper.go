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

	"github.com/apprehensions/rbxbin"
	"github.com/godbus/dbus/v5"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/studiorpc"
)

const killWait = 3 * time.Second

const (
	// Randomly chosen log entry in cases where Studios process
	// continues to run. Due to a lack of bug reports, it is unknown
	// specifically which entry to use for these types of cases.
	shutdownEntry = "[FLog::LifecycleManager] Exited ApplicationScope"
)

type bootstrapper struct {
	*ui
	win adw.Window

	pbar   gtk.ProgressBar
	status gtk.Label

	dir string
	bin rbxbin.Deployment

	rp *studiorpc.StudioRPC
}

func (s *ui) newBootstrapper() *bootstrapper {
	builder := gtk.NewBuilderFromResource("/org/vinegarhq/Vinegar/ui/bootstrapper.ui")
	defer builder.Unref()

	b := bootstrapper{
		ui: s,
		rp: studiorpc.New(),
	}

	builder.GetObject("window").Cast(&b.win)
	b.win.SetApplication(&s.app.Application)
	s.app.AddWindow(&b.win.Window)
	destroy := func(_ gtk.Window) bool {
		// https://github.com/jwijenbergh/puregotk/issues/17
		// BUG: realistically no other way to cancel
		//      the bootstrapper, so just exit immediately!
		b.app.Quit()
		return false
	}
	b.win.ConnectCloseRequest(&destroy)

	builder.GetObject("status").Cast(&b.status)
	builder.GetObject("progress").Cast(&b.pbar)
	b.status.Unref()
	b.pbar.Unref()

	b.win.Present()
	b.win.Unref()

	return &b
}

func (b *bootstrapper) performing() func() {
	var tcb glib.SourceFunc
	tcb = func(uintptr) bool {
		b.pbar.Pulse()
		return true
	}
	id := glib.TimeoutAdd(128, &tcb, null)
	return func() { glib.SourceRemove(id) }
}

func (b *bootstrapper) message(msg string, args ...any) {
	slog.Info(msg, args...)
	idle(func() { b.status.SetLabel(msg + "...") })
}

func (b *bootstrapper) start() error {
	return b.run()
}

func (b *bootstrapper) run(args ...string) error {
	if err := b.setup(); err != nil {
		return fmt.Errorf("setup: %w", err)
	}

	if err := b.execute(args...); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
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

func (b *bootstrapper) handleRobloxLog(line string) {
	if !b.cfg.Studio.Quiet {
		slog.Log(context.Background(), logging.LevelRoblox, line)
	}

	if strings.Contains(line, shutdownEntry) {
		go func() {
			time.Sleep(killWait)
			syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()
	}

	if b.cfg.Studio.DiscordRPC {
		if err := b.rp.Handle(line); err != nil {
			slog.Error("Presence handling failed", "error", err)
		}
	}
}

func (b *bootstrapper) registerGameMode(pid int32) {
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

func (b *bootstrapper) setupMIME() error {
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
