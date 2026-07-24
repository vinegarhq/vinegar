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

	if err := b.setupMimalloc(); err != nil {
		return fmt.Errorf("mimalloc: %w", err)
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

// setupMimalloc makes Studio's mimalloc-redirect functional under Wine;
// mimalloc deployments hard-assert at startup without it. The redirect
// can only patch the real ucrtbase (Wine's builtin lacks the private
// _expand_base/_recalloc_base/_msize_base exports it resolves), and Wine
// pins ucrtbase to the KnownDlls copy in system32 — the application
// directory is never searched — so the deployment's native DLL is copied
// there and enabled for Studio alone via AppDefaults. A session-wide
// WINEDLLOVERRIDES would instead break wineboot and every other system
// process, whose ucrtbase must stay builtin. The env side of this
// accommodation (MIMALLOC_FORCE_REDIRECT et al.) lives in config.Prefix.
func (b *bootstrapper) setupMimalloc() error {
	system32 := filepath.Join(dirs.Prefixes, "studio", "drive_c", "windows", "system32")
	dst := filepath.Join(system32, "ucrtbase.dll")
	backup := dst + ".wine-builtin"
	override := `HKCU\Software\Wine\AppDefaults\RobloxStudioBeta.exe\DllOverrides`

	if _, err := os.Stat(filepath.Join(b.dir, "ucrtbase.dll")); err != nil {
		// Deployment ships no native CRT: undo a previous setup so the
		// native-only override cannot point at Wine's builtin file.
		if _, err := os.Stat(backup); err != nil {
			return nil
		}
		if err := os.Rename(backup, dst); err != nil {
			return err
		}
		return b.pfx.RegistryAdd(override, "ucrtbase", "builtin")
	}

	b.message(L("Setting up mimalloc"))

	if _, err := os.Stat(backup); errors.Is(err, os.ErrNotExist) {
		if err := cp.Copy(dst, backup); err != nil {
			return err
		}
	}
	if err := cp.Copy(filepath.Join(b.dir, "ucrtbase.dll"), dst); err != nil {
		return err
	}
	return b.pfx.RegistryAdd(override, "ucrtbase", "native")
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
