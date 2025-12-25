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

	. "github.com/pojntfx/go-gettext/pkg/i18n"
)

func (b *bootstrapper) setup() error {
	if len(b.procs) > 0 {
		slog.Info("Skipping setup!", "ver", b.bin.GUID)
		return nil
	}

	firstSetup := !b.pfx.Exists()
	if err := b.prepareWine(); err != nil {
		return err
	}

	// Don't bother retrieving the security if the wineprefix
	// was initialized just now
	if b.rbx.Security == "" && !firstSetup {
		stop := b.performing()
		b.message(L("Acquiring user authentication"))
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

	if err := b.updateDeployment(); err != nil {
		return err
	}

	if err := b.preRun(); err != nil {
		return err
	}

	return nil
}

func (b *bootstrapper) copyOverlay() error {
	dir := filepath.Join(dirs.Overlays, strings.ToLower(studio.Short()))

	// Don't copy Overlay if it doesn't exist
	_, err := os.Stat(dir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	b.message(L("Copying Overlay"))

	return cp.Copy(dir, b.dir)
}

func (b *bootstrapper) preRun() error {
	defer b.performing()()

	if err := b.copyOverlay(); err != nil {
		return fmt.Errorf("overlay: %w", err)
	}

	if err := b.applyFFlags(); err != nil {
		return fmt.Errorf("fflags: %w", err)
	}

	gutil.IdleAdd(func() { b.status.SetLabel(L("Launching Studio")) })

	dpi := 96.0 * b.win.GetNative().GetSurface().GetScale()
	slog.Info("Updating Wine DPI", "dpi", dpi)
	if err := b.pfx.RegistryAdd(`HKEY_CURRENT_USER\Control Panel\Desktop`,
		"LogPixels", uint32(dpi)); err != nil {
		return fmt.Errorf("scale set: %w, err")
	}

	return nil
}

func (b *bootstrapper) applyFFlags() error {
	// This actually includes DXVK.
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

	b.message(L("Applying FFlags"))
	return f.Apply(b.dir)
}
