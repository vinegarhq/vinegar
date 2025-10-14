package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"syscall"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/sewnie/rbxbin"
	"github.com/vinegarhq/vinegar/internal/gtkutil"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/studiorpc"
)

type bootstrapper struct {
	*app
	win adw.Window

	pbar   gtk.ProgressBar
	status gtk.Label
	info   gtk.Label

	dir string
	bin *rbxbin.Deployment

	procs []*os.Process

	rp *studiorpc.StudioRPC
}

func (a *app) newBootstrapper() *bootstrapper {
	builder := gtk.NewBuilderFromResource(gtkutil.Resource("ui/bootstrapper.ui"))
	defer builder.Unref()

	b := bootstrapper{
		app: a,
		rp:  studiorpc.New(),
	}

	builder.GetObject("window").Cast(&b.win)
	destroy := func(_ gtk.Window) bool {
		// TODO: context cancellation
		b.Quit()
		return false
	}
	b.win.ConnectCloseRequest(&destroy)

	builder.GetObject("status").Cast(&b.status)
	builder.GetObject("progress").Cast(&b.pbar)
	builder.GetObject("info").Cast(&b.info)
	b.status.Unref()
	b.pbar.Unref()

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
	gtkutil.IdleAdd(func() { b.status.SetLabel(msg) })
}

func (b *bootstrapper) run(args ...string) error {
	if b.win.GetApplication() != nil && b.win.IsVisible() {
		slog.Warn("Bootstrapper currently in setup, ignoring run request")
		return nil
	}

	gtkutil.IdleAdd(func() {
		b.app.AddWindow(&b.win.Window)
		b.win.Present()
	})
	defer gtkutil.IdleAdd(func() {
		b.app.RemoveWindow(&b.win.Window)
		b.win.SetVisible(false) // Incase bailed out
	})

	if err := b.setup(); err != nil {
		return fmt.Errorf("setup: %w", err)
	}

	return b.execute(args...)
}

func (b *bootstrapper) handleRobloxLog(line string) {
	switch {
	case strings.Contains(line, "ANR In Progress. ApplicationState: Background"):
		// Roblox normally exits after submitting ANR data to
		// ecsv2.roblox.com, but in Wine it does nothing.
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		gtkutil.IdleAdd(func() {
			b.showError(errors.New(
				"Studio detected as unresponsive!"))
		})
	case strings.Contains(line, "LoginDialog Error: Embedded Web Browser fail to load"):
		// Ensure that browser login functionality will work
		if err := b.setMime(); err != nil {
			gtkutil.IdleAdd(func() {
				b.pfx.Kill()
				b.showError(err)
			})
		}
	}

	if b.cfg.Studio.DiscordRPC {
		if err := b.rp.Handle(line); err != nil {
			slog.Error("Presence handling failed", "error", err)
		}
	}

	// 2025-08-17T13:13:37.469Z,11.469932,0238,6,Info [FLog::AnrDetector]
	// 2025-08-17T12:54:23.583Z,1.583294,00e0,6(,(Warning|Info|Error)) [Flog::..] ...
	_, a, ok := strings.Cut(line, ",6")
	if ok {
		i := strings.Index(a, " [")
		if i > 0 && a[1:i] == "Info" && !b.cfg.Debug {
			return
		}
		line = strings.TrimSpace(a[i:])
	}

	slog.Log(context.Background(), logging.LevelRoblox.Level(), line)
}

func (b *bootstrapper) registerGameMode(target int) error {
	if !b.cfg.Studio.GameMode || b.bus == nil {
		return nil
	}

	resp, err := b.bus.CallSync("org.freedesktop.portal.Desktop",
		"/org/freedesktop/portal/desktop",
		"org.freedesktop.portal.GameMode",
		"RegisterGame",
		glib.NewVariant("(i)", target),
		glib.NewVariantType("(i)"),
		gio.GDbusCallFlagsNoneValue,
		-1,
		nil,
	)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	var res int32
	resp.Get("(i)", &res)

	if res < 0 {
		return errors.New("rejected by gamemode")
	}
	slog.Info("Registered with GameMode", "response", res)

	return nil
}
