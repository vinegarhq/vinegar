package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/sewnie/rbxbin"
	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/studiorpc"
)

var (
	studioRegPath = `HKCU\Software\Roblox\RobloxStudio`
	backupExport  = filepath.Join(dirs.Data, "settings.reg")
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
	builder := gtk.NewBuilderFromResource(gutil.Resource("ui/bootstrapper.ui"))
	defer builder.Unref()

	b := bootstrapper{
		app: a,
		rp:  studiorpc.New(a.rbx),
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
	gutil.IdleAdd(func() { b.status.SetLabel(msg) })
}

func (b *bootstrapper) run(args ...string) error {
	if b.win.GetApplication() != nil && b.win.IsVisible() {
		slog.Warn("Bootstrapper currently in setup, ignoring run request")
		return nil
	}

	gutil.IdleAdd(func() {
		b.app.AddWindow(&b.win.Window)
		b.win.Present()
	})
	defer gutil.IdleAdd(func() {
		b.app.RemoveWindow(&b.win.Window)
		b.win.SetVisible(false) // Incase bailed out
	})

	if err := b.setup(); err != nil {
		return fmt.Errorf("setup: %w", err)
	}

	err := b.execute(args...)

	if err := b.backupSettings(); err != nil {
		slog.Error("Failed to backup Studio settings", "err", err)
	}

	return err
}

func (b *bootstrapper) restoreSettings() error {
	if !b.pfx.Exists() {
		return nil
	}
	_, err := os.Stat(backupExport)
	if err != nil {
		return nil
	}

	slog.Info("Restoring Studio settings")
	return b.pfx.RegistryImportFile(backupExport)
}

func (b *bootstrapper) backupSettings() error {
	if err := b.prepareWine(); err != nil {
		return err
	}
	if !b.pfx.Running() {
		return nil
	}

	// Studio stores its layout information as a multiline XML
	// in the registry. This is problematic for [Prefix.RegistryQuery]
	// as it is infeasible to parse. Grab the entire registry instead.

	r, err := b.pfx.Registry()
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	stripped := wine.NewRegistryKey(`HKCU\`)
	s := r.Query(studioRegPath)
	if s == nil {
		// were to be added to a key, HKCU gets deleted :p
		slog.Warn("Nothing to backup!")
		return nil
	}

	stripped.AddKey(s)
	if s := r.Query(credRegPath); s != nil {
		stripped.AddKey(s)
	}

	f, err := os.OpenFile(backupExport,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("file: %w", err)
	}

	if err := stripped.Export(f); err != nil {
		return err
	}

	slog.Debug("Backed up Studio Registry")

	return nil
}

func (b *bootstrapper) handleRobloxLog(line string) {
	switch {
	case strings.Contains(line, "LoginDialog Error: Embedded Web Browser fail to load"):
		// Ensure that browser login functionality will work
		if err := b.setMime(); err != nil {
			gutil.IdleAdd(func() {
				b.pfx.Kill()
				b.showError(err)
			})
		}
	}

	// time,runtime,code,code2[,level ] ...
	{
		entry := strings.SplitN(line, ",", 4)
		if len(entry) < 3 {
			slog.Log(context.Background(), slog.LevelInfo, line)
			return
		}
		entry = entry[3:]
		if len(entry) != 1 {
			panic(entry)
		}
		line = entry[0]
	}

	i := strings.Index(line, " [")
	level := "Info"
	if i > 0 {
		if j := strings.Index(line[:i], ","); j > 0 {
			level = line[j+1 : i]
		}
		line = line[i+1:]
	}

	if b.cfg.Studio.DiscordRPC {
		if err := b.rp.Handle(line); err != nil {
			slog.Error("Discord Rich Presence handling failed", "err", err)
		}
	}

	if level == "Info" && !b.cfg.Debug {
		return
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
