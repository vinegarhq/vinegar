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

	"github.com/godbus/dbus/v5"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/sewnie/rbxbin"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/studiorpc"
	"github.com/vinegarhq/vinegar/sysinfo"
)

type bootstrapper struct {
	*app
	win adw.Window

	pbar   gtk.ProgressBar
	status gtk.Label
	info   gtk.Label

	dir string
	bin *rbxbin.Deployment

	rp *studiorpc.StudioRPC
}

func (s *app) newBootstrapper() *bootstrapper {
	builder := gtk.NewBuilderFromResource("/org/vinegarhq/Vinegar/ui/bootstrapper.ui")
	defer builder.Unref()

	b := bootstrapper{
		app: s,
		rp:  studiorpc.New(),
	}

	builder.GetObject("window").Cast(&b.win)
	s.AddWindow(&b.win.Window)
	destroy := func(_ gtk.Window) bool {
		// https://github.com/jwijenbergh/puregotk/issues/17
		// BUG: realistically no other way to cancel
		//      the bootstrapper, so just exit immediately!
		b.Quit()
		return false
	}
	b.win.ConnectCloseRequest(&destroy)

	builder.GetObject("status").Cast(&b.status)
	builder.GetObject("progress").Cast(&b.pbar)
	builder.GetObject("info").Cast(&b.info)
	b.status.Unref()
	b.pbar.Unref()

	b.win.Show()

	return &b
}

func (b *bootstrapper) performing() func() {
	var tcb glib.SourceFunc = func(uintptr) bool {
		b.pbar.Pulse()
		return true
	}
	id := glib.TimeoutAdd(128, &tcb, 0)
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
		before, ent, found := strings.Cut(line, ",6")
		if !found {
			ent = before
		} else if ent[0] == ',' || ent[0] == ' ' {
			ent = ent[1:]
		}
		slog.Log(context.Background(), logging.LevelRoblox, ent)
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
	// xdg-mime unavailable in Flatpak; Flatpak
	// handles MIME associations for us.
	if sysinfo.InFlatpak {
		return nil
	}

	o, err := exec.Command("xdg-mime", "default",
		"org.vinegarhq.Vinegar.studio.desktop",
		"x-scheme-handler/roblox-studio",
		"x-scheme-handler/roblox-studio-auth",
		"application/x-roblox-rbxl",
		"application/x-roblox-rbxlx",
	).CombinedOutput()
	if err == nil {
		return nil
	}

	if len(o) > 0 {
		return fmt.Errorf("setup mime: %s", string(o))
	}
	return fmt.Errorf("xdg-mime: %w", err)
}
