package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gtkutil"
	"github.com/vinegarhq/vinegar/internal/logging"
)

func (ctl *control) hideRunUntil() func() {
	var button gtk.Button
	var stack gtk.Stack
	ctl.builder.GetObject("stack").Cast(&stack)
	ctl.builder.GetObject("btn-run").Cast(&button)
	stack.SetVisibleChildName("stkpage-spinner")
	button.SetSensitive(false)
	return func() {
		gtkutil.IdleAdd(func() {
			button.SetSensitive(true)
			stack.SetVisibleChildName("stkpage-btn")
			ctl.updateRun()
		})
	}
}

func (ctl *control) run() error {
	show := ctl.hideRunUntil()
	defer show()

	if ctl.pfx.Exists() && len(ctl.boot.procs) == 0 {
		if err := ctl.app.boot.setup(); err != nil {
			return fmt.Errorf("setup: %w", err)
		}

		// When command is finally executed
		h := ctl.app.boot.win.ConnectSignal("notify::visible", &show)
		defer func() { ctl.app.boot.win.DisconnectSignal(h) }()
		return ctl.app.boot.run()
	}

	if len(ctl.boot.procs) > 0 {
		return ctl.pfx.Kill()
	}
	return ctl.app.boot.setupPrefix()
}

// No error will be returned, error handling is displayed
// from an AdwBanner.
func (ctl *control) saveConfig() {
	var banner adw.Banner
	ctl.builder.GetObject("banner-config-error").Cast(&banner)
	banner.SetRevealed(false)

	if err := ctl.cfg.Save(); err != nil {
		ctl.showError(err)
		return
	}

	err := ctl.reload()
	if err == nil {
		return
	}

	banner.SetTitle(fmt.Sprintf("Invalid: %v", err))
	banner.SetRevealed(true)
	slog.Info("Configuration validation error", "err", err)
	return
}

func (ctl *control) showAbout() {
	w := adw.NewAboutDialogFromAppdata(gtkutil.Resource("metainfo.xml"), version[1:])
	w.Present(&ctl.win.Widget)
	w.Unref()
}

func (ctl *control) deleteDeployments() error {
	return os.RemoveAll(dirs.Versions)
}

func (ctl *control) deletePrefixes() error {
	defer ctl.hideRunUntil()()
	slog.Info("Deleting Wineprefixes!")

	// Wineserver isn't required if it's missing.
	_ = ctl.pfx.Kill()

	return os.RemoveAll(dirs.Prefixes)
}

func (ctl *control) clearCache() error {
	return filepath.WalkDir(dirs.Cache, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if path == dirs.Cache || path == dirs.Logs || path == logging.Path {
			return nil
		}

		slog.Info("Removing cache file", "path", path)
		return os.RemoveAll(path)
	})
}
