package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/sewnie/wine/dxvk"
	"github.com/sewnie/wine/peutil"
	"github.com/sewnie/wine/webview2"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
)

func (b *bootstrapper) setupPrefix() error {
	defer b.performing()()
	b.message("Setting up Wine")

	_, err := os.Stat(dirs.WinePath)
	if b.cfg.Studio.WineRoot.IsDefault() && err != nil {
		if err := b.updateWine(); err != nil {
			return fmt.Errorf("dl: %w", err)
		}
	}

	// Handles Wineprefix initialization as necessary
	if err := b.pfx.Start(); err != nil {
		return err
	}

	// Latest versions of studio require a implemented call, check if the given
	// prefix supports it
	if b.cfg.Studio.ForcedVersion != "" {
		// Skip check on old versions, which will cause the user to remove the override,
		// and get a proper error afterwards :)
		return nil
	}

	b.message("Checking Wineprefix")

	f, err := peutil.Open(filepath.Join(
		b.pfx.Dir(), "drive_c", "windows", "system32", "kernelbase.dll"))
	if err != nil {
		// WINE probably won't change this path anytime soon, so this DLL being missing
		// is catastrophic
		return fmt.Errorf("dll: %w", err)
	}
	defer f.Close()

	es, err := f.Exports()
	if err != nil {
		return fmt.Errorf("exports: %w", err)
	}

	if !slices.ContainsFunc(es, func(e peutil.Export) bool {
		return e.Name == "VirtualProtectFromApp"
	}) {
		return errors.New("Wine installation cannot run studio; update wine to >=10.13")
	}

	return nil
}

func (b *bootstrapper) setupDxvk() error {
	if !b.cfg.Studio.Renderer.IsDXVK() {
		return nil
	}

	// If DXVK is installed in the wineprefix, uninstallation
	// won't be necessary if it's disabled as it still requires
	// DLL overrides to be present.

	new := b.cfg.Studio.Renderer.DXVKVersion()
	b.message("Checking DXVK", "version", new)

	current, err := dxvk.Version(b.pfx)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	if current == new {
		return nil
	}
	b.message("Downloading DXVK", "current", current, "new", new)

	name := filepath.Join(dirs.Cache, "dxvk-"+new+".tar.gz")
	if _, err := os.Stat(name); err == nil {
		goto install
	}

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return fmt.Errorf("prepare cache: %w", err)
	}

	if err := netutil.DownloadProgress(
		dxvk.URL(new), name, &b.pbar); err != nil {
		return fmt.Errorf("download: %w", err)
	}

install:
	defer b.performing()()

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	b.message("Extracting DXVK", "version", new)

	if err := dxvk.Extract(b.pfx, f); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}

func (b *bootstrapper) setupWebView() error {
	new := b.cfg.Studio.WebView.String()

	installer := filepath.Join(dirs.Cache, "webview-"+new+".exe")
	b.message("Checking WebView", "version", new)

	current := webview2.Current(b.pfx)
	if current != "" && current != new {
		b.message("Uninstalling WebView", "current", current, "new", new)
		if err := webview2.Uninstall(b.pfx, current); err != nil {
			return fmt.Errorf("uninstall: %w", err)
		}
	}
	if current == new || b.cfg.Studio.WebView.Enabled() {
		return nil
	}

	if _, err := os.Stat(installer); err != nil {
		stop := b.performing()
		b.message("Fetching WebView", "upload", b.cfg.Studio.WebView)
		webview2.Client.Transport.(*http.Transport).DisableCompression = true
		d, err := webview2.Stable.Runtime(new, "x64")
		if err != nil {
			return fmt.Errorf("fetch: %w", err)
		}
		stop()

		b.message("Downloading WebView", "catalog", d.Delivery.CatalogID)
		if err := netutil.DownloadProgress(d.URL, installer, &b.pbar); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	b.message("Installing WebView", "path", installer)
	defer b.performing()()

	return webview2.Install(b.pfx, installer)
}
