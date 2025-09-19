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

	pfxFirstRun := !b.pfx.Exists()

	if err := b.setupPrefix(); err != nil {
		return err
	}

	if b.rbx.Security == "" && !pfxFirstRun {
		b.message("Acquiring user authentication")
		if err := b.app.getSecurity(); err != nil {
			slog.Warn("Retrieving authenticated user failed", "err", err)
		}
	}

	if err := b.stepSetupDxvk(); err != nil {
		return fmt.Errorf("dxvk dl: %w", err)
	}

	if err := b.stepWebviewDownload(); err != nil {
		return fmt.Errorf("webview dl: %w", err)
	}

	if err := b.setupDeployment(); err != nil {
		return err
	}

	if err := b.stepWebviewInstall(); err != nil {
		return fmt.Errorf("webview: %w", err)
	}

	if err := b.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
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

	uiThread(func() { b.status.SetLabel("Launching Studio") })

	// If no setup took place, this will go immediately.
	slog.Info("Kickstarting Wineserver")
	if err := b.pfx.Server(); err != nil {
		return fmt.Errorf("server: %w, check logs", err)
	}

	// Running this command will initialize Wine for
	// running applications, which gives more time to the
	// splash window to show that studio is going to be ran.
	if err := b.stepChangeStudioTheme(); err != nil {
		slog.Warn("Failed to change Studio's theme!", "err", err)
	}

	return nil
}

func (b *bootstrapper) stepChangeStudioTheme() error {
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
	b.message("Changing Theme", "theme", theme)
	return b.pfx.RegistryAdd(key, val, theme)
}
