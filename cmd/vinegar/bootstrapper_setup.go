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
	"github.com/vinegarhq/vinegar/internal/dirs"

	. "github.com/pojntfx/go-gettext/pkg/i18n"
)

// Ideally:
//   - WebView version and ROBLOSECURITY should be checked only
//     once by using the offline registry
//   - Download Webview as neccesary
//   - Install Roblox
//   - Install DXVK
//   - Install WebView
func (b *bootstrapper) setupExecute() error {
	if b.count > 0 {
		slog.Info("Skipping setup!", "ver", b.bin.GUID)
		return nil
	}

	// If the registry does not exist, the Wineprefix has not been
	// initialized yet, which makes things such as Webview, DXVK
	// as uninstalled and the ROBLOSECURITY as missing.
	offline, err := b.pfx.Registry()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("registry: %w", err)
	}

	webview := b.webViewVersion(offline)

	if err := b.getSecurity(offline); err != nil {
		slog.Warn("Retrieving authenticated user failed", "err", err)
	}

	if err := b.updateDeployment(); err != nil {
		return err
	}

	// These tasks are so fast, a performing indicator
	// is not going to be necessary.
	if err := b.copyOverlay(); err != nil {
		return fmt.Errorf("overlay: %w", err)
	}

	if err := b.applyFFlags(); err != nil {
		return fmt.Errorf("fflags: %w", err)
	}

	// Does nothing if WebView is disabled, preferred to download
	// a large installer before Wineprefix initialization.
	// before Wineprefix initialization.
	if err := b.downloadWebView(webview); err != nil {
		return fmt.Errorf("download webview: %w", err)
	}

	stop := b.performing()
	defer stop()

	// Initializes the wineprefix and wineserver, along with
	// setting up Vinegar's registry values as necessary, along
	// with restoring settings. After the wineserver is ran,
	// it is safe to install DXVK, WebView, and other modifications.
	if _, err := b.app.prepareWine(); err != nil {
		return err
	}

	stop()

	if err := b.installWebView(webview); err != nil {
		return fmt.Errorf("install webview: %w", err)
	}

	// Currently, DXVK does not quite invoke any sort of application,
	// giving the wineserver the persistent timeout until another program
	// is executed. Attempt to reduce chances of being killed by installing
	// DXVK only before running Studio, which leaves the server open.
	if err := b.setupDXVK(); err != nil {
		return fmt.Errorf("dxvk: %w", err)
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

func (b *bootstrapper) applyFFlags() error {
	f := maps.Clone(b.cfg.Studio.FFlags)

	if r := b.cfg.Studio.Renderer; r != "" {
		renderers := []string{"D3D11", "Vulkan", "D3D11FL10", "OpenGL"}
		if v := b.cfg.Studio.DXVKVersion(); v != "" {
			r = "D3D11"
		}

		if !slices.Contains(renderers, string(r)) {
			return fmt.Errorf("unknown renderer: %s", r)
		}

		for _, avail := range renderers {
			isRenderer := avail == string(r)
			f["FFlagDebugGraphicsPrefer"+avail] = isRenderer
			f["FFlagDebugGraphicsDisable"+avail] = !isRenderer
		}
	}

	b.message(L("Applying FFlags"))
	return f.Apply(b.dir)
}
