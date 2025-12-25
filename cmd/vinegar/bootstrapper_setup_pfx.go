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

	. "github.com/pojntfx/go-gettext/pkg/i18n"
)

func (b *bootstrapper) prepareWine() error {
	defer b.performing()()

	// Handles Wineprefix initialization as necessary
	if _, err := b.app.prepareWine(); err != nil {
		return err
	}

	// Latest versions of studio require a implemented call, check if the given
	// prefix supports it
	if b.cfg.Studio.ForcedVersion != "" {
		// Skip check on old versions, which will cause the user to remove the override,
		// and get a proper error afterwards :)
		return nil
	}

	b.message(L("Checking Wineprefix"))

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

	version := b.cfg.Studio.Renderer.DXVKVersion()
	b.message(L("Checking DXVK"), "against", version)

	installed, err := dxvk.Version(b.pfx)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	if installed == version {
		return nil
	}
	b.message(L("Downloading DXVK"), "current", installed, "new", version)

	name := filepath.Join(dirs.Cache, "dxvk-"+version+".tar.gz")
	if _, err := os.Stat(name); err == nil {
		goto install
	}

	if err := os.MkdirAll(dirs.Cache, 0o755); err != nil {
		return fmt.Errorf("prepare cache: %w", err)
	}

	if err := netutil.DownloadProgress(
		dxvk.URL(version), name, &b.pbar); err != nil {
		return fmt.Errorf("download: %w", err)
	}

install:
	defer b.performing()()

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	b.message(L("Extracting DXVK"), "version", version)

	if err := dxvk.Extract(b.pfx, f); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}

func (b *bootstrapper) setupWebView() error {
	version := b.cfg.Studio.WebView.String()

	installer := filepath.Join(dirs.Cache, "webview-"+version+".exe")
	b.message(L("Checking WebView"), "against", version)

	installed := webview2.Current(b.pfx)
	if installed != "" && installed != version {
		b.message(L("Uninstalling WebView"), "current", installed, "new", version)
		if err := webview2.Uninstall(b.pfx, installed); err != nil {
			return fmt.Errorf("uninstall: %w", err)
		}
	}
	if installed == version || b.cfg.Studio.WebView.Enabled() {
		return nil
	}

	if _, err := os.Stat(installer); err != nil {
		stop := b.performing()
		b.message(L("Fetching WebView"), "upload", b.cfg.Studio.WebView)
		// Microsoft doesn't like compressed requests
		webview2.Client.Transport.(*http.Transport).DisableCompression = true
		d, err := webview2.Stable.Runtime(version, "x64")
		if err != nil {
			return fmt.Errorf("fetch: %w", err)
		}
		stop()

		b.message(L("Downloading WebView"), "catalog", d.Delivery.CatalogID)
		if err := netutil.DownloadProgress(d.URL, installer, &b.pbar); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	b.message(L("Installing WebView"), "version", version, "path", installer)
	defer b.performing()()

	return webview2.Install(b.pfx, installer)
}
