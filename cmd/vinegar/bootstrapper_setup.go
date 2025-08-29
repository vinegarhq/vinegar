package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	cp "github.com/otiai10/copy"
	"github.com/vinegarhq/vinegar/internal/dirs"
)

func (b *bootstrapper) setup() error {
	b.removePlayer()

	// Bootstrapper is currently running
	if b.bin != nil {
		slog.Info("Skipping setup!", "ver", b.bin.GUID)
		return nil
	}

	if err := b.setupDeployment(); err != nil {
		return err
	}

	if err := b.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	if err := b.setupPrefix(); err != nil {
		return fmt.Errorf("prefix: %w", err)
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

	b.message("Renderer Status:", "renderer", b.cfg.Studio.Renderer,
		"dxvk", b.cfg.Studio.Dxvk)

	if err := b.setupOverlay(); err != nil {
		return fmt.Errorf("setup overlay: %w", err)
	}

	b.message("Applying FFlags")
	if err := b.cfg.Studio.FFlags.Apply(b.dir); err != nil {
		return fmt.Errorf("apply fflags: %w", err)
	}

	// When Studio is finally executed, the bootstrapper window
	// immediately closes. Indicate to the user that Studio is being ran.
	idle(func() { b.status.SetLabel("Launching Studio") })

	// To allow the wineserver to report any errors, start it manually,
	// instead of have it appear in any registry calls.
	slog.Info("Kickstarting Wineserver")
	if err := b.pfx.Server(); err != nil {
		return fmt.Errorf("server: %w, check logs", err)
	}

	if err := b.changeTheme(); err != nil {
		slog.Warn("Failed to change Studio's theme!", "err", err)
	}

	return nil
}

func (b *bootstrapper) changeTheme() error {
	key := `HKEY_CURRENT_USER\Software\Roblox\RobloxStudio\Themes`
	val := "CurrentTheme"
	q, err := b.pfx.RegistryQuery(key, val)
	if err != nil {
		slog.Warn("Failed to retrieve current Studio theme", "err", err)
		return nil
	}

	// If the user set an explicit theme rather than the default
	// Studio system theme, do not attempt to change the theme accordingly.
	if len(q) != 0 && q[0].Subkeys[0].Value != "Default" {
		return nil
	}

	theme := "Light"
	if b.GetStyleManager().GetDark() {
		theme = "Dark"
	}
	slog.Info("Changing Studio Theme", "theme", theme)
	return b.pfx.RegistryAdd(key, val, theme)
}
