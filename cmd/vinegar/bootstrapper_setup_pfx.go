package main

import (
	"errors"
	"fmt"
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

	// Always initialize in case Wine changes,
	// to prevent a dialog from appearing in normal apps.
	b.message("Initializing Wineprefix", "dir", b.pfx.Dir())
	if err := b.pfx.Init().Run(); err != nil {
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

func (b *bootstrapper) stepSetupDxvk() error {
	// If DXVK is installed in the wineprefix, uninstallation
	// won't be necessary if it's disabled as it still requires
	// DLL overrides to be present.
	if b.cfg.Studio.DXVK == "" {
		return nil
	}

	b.message("Checking DXVK", "version", b.cfg.Studio.DXVK)

	new := string(b.cfg.Studio.DXVK)
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
		dxvk.URL(b.cfg.Studio.DXVK.String()), name, &b.pbar); err != nil {
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

func (b *bootstrapper) webviewInstaller() string {
	if b.cfg.Studio.Webview == "" {
		return ""
	}
	return filepath.Join(dirs.Cache, "webview-"+b.cfg.Studio.Webview+".exe")
}

func (b *bootstrapper) webviewPath() string {
	return filepath.Join(b.pfx.Dir(), "drive_c/Program Files (x86)/Microsoft/EdgeWebView/Application", b.cfg.Studio.Webview)
}

func (b *bootstrapper) stepWebviewDownload() error {
	name := b.webviewInstaller()
	if name == "" {
		return nil
	}

	if _, err := os.Stat(name); err == nil {
		return nil
	}

	stop := b.performing()
	b.message("Fetching WebView", "upload", b.cfg.Studio.Webview)
	d, err := webview2.StableLegacy.Runtime(b.cfg.Studio.Webview, "x64")
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	stop()

	b.message("Downloading WebView", "catalog", d.Delivery.CatalogID)
	return netutil.DownloadProgress(d.URL, name, &b.pbar)
}

func (b *bootstrapper) stepWebviewInstall() error {
	name := b.webviewInstaller()
	path := b.webviewPath()

	_, err := os.Stat(path)
	if err == nil && name == "" {
		b.message("Uninstalling WebView")

		return os.RemoveAll(path)
	} else if name == "" || (err == nil && name != "") {
		return nil
	}

	b.message("Installing WebView", "path", name)
	defer b.performing()()

	return webview2.Install(b.pfx, name)
}
