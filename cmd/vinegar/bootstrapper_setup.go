package main

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	cp "github.com/otiai10/copy"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
)

func (b *bootstrapper) setup() error {
	if len(b.procs) > 0 {
		slog.Info("Skipping setup!", "ver", b.bin.GUID)
		return nil
	}

	pfxFirstRun := !b.pfx.Exists()

	if err := b.setupPrefix(); err != nil {
		return err
	}

	if b.rbx.Security == "" && !pfxFirstRun {
		stop := b.performing()
		b.message("Acquiring user authentication")
		if err := b.app.getSecurity(); err != nil {
			slog.Warn("Retrieving authenticated user failed", "err", err)
		}
		stop()
	}

	if err := b.setupDxvk(); err != nil {
		return fmt.Errorf("dxvk: %w", err)
	}

	if err := b.setupWebView(); err != nil {
		return fmt.Errorf("webview: %w", err)
	}

	if err := b.setupDeployment(); err != nil {
		return err
	}

	if err := b.stepPrepareRun(); err != nil {
		return err
	}

	return nil
}

func (b *bootstrapper) setupOverlay() error {
	dir := filepath.Join(dirs.Overlays, strings.ToLower(studio.Short()))

	// Don't copy Overlay if it doesn't exist
	_, err := os.Stat(dir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	b.message("Copying Overlay")

	return cp.Copy(dir, b.dir)
}

func (b *bootstrapper) stepPrepareRun() error {
	defer b.performing()()

	if err := b.setupOverlay(); err != nil {
		return fmt.Errorf("setup overlay: %w", err)
	}

	if err := b.stepApplyFFlags(); err != nil {
		return fmt.Errorf("fflags: %w", err)
	}

	gutil.IdleAdd(func() { b.status.SetLabel("Launching Studio") })

	// The following registry modifications starts and prepares Wine.
	slog.Info("Kickstarting Wineserver")

	dpi := 96.0 * b.win.GetNative().GetSurface().GetScale()
	slog.Info("Updating Wine DPI", "dpi", dpi)
	if err := b.pfx.RegistryAdd(`HKEY_CURRENT_USER\Control Panel\Desktop`,
		"LogPixels", uint32(dpi)); err != nil {
		return fmt.Errorf("scale set: %w, err")
	}

	return nil
}

func (b *bootstrapper) stepApplyFFlags() error {
	renderers := config.Renderer("").Values()

	f := maps.Clone(b.cfg.Studio.FFlags)
	if b.cfg.Studio.Renderer != "" {
		if !slices.Contains(renderers, string(b.cfg.Studio.Renderer)) {
			return fmt.Errorf("unknown renderer: %s", b.cfg.Studio.Renderer)
		}

		for _, r := range renderers {
			isRenderer := r == string(b.cfg.Studio.Renderer)
			f["FFlagDebugGraphicsPrefer"+r] = isRenderer
			f["FFlagDebugGraphicsDisable"+r] = !isRenderer
		}
	}

	b.message("Applying FFlags")
	if err := f.Apply(b.dir); err != nil {
		return fmt.Errorf("apply fflags: %w", err)
	}

	return nil
}
