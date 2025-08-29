package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/sewnie/wine/dxvk"
	"github.com/sewnie/wine/peutil"
	"github.com/sewnie/wine/webview"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
)

func (b *bootstrapper) setupPrefix() error {
	b.message("Setting up Wine")

	// if Wine is runnable
	if c := b.pfx.Wine(""); c.Err != nil {
		return fmt.Errorf("wine: %w", c.Err)
	}

	// Always initialize in case Wine changes,
	// to prevent a dialog from appearing in normal apps.
	if err := b.stepPrefixInit(); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	if err := b.checkPrefix(); err != nil {
		return err
	}

	return nil
}

func (b *bootstrapper) stepPrefixInit() error {
	defer b.performing()()

	b.message("Initializing Wineprefix", "dir", b.pfx.Dir())
	return run(b.pfx.Init())
}

func (b *bootstrapper) checkPrefix() error {
	b.message("Checking Wineprefix")
	// Latest versions of studio require a implemented call, check if the given
	// prefix supports it
	if b.cfg.Studio.ForcedVersion != "" {
		// Skip check on old versions, which will cause the user to remove the override,
		// and get a proper error afterwards :)
		return nil
	}

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
		// TODO: actually give a solution to the user
		return errors.New("wine installation cannot run studio")
	}

	return nil
}

func (b *bootstrapper) stepSetupDxvk() error {
	// If installed, installing won't be required since
	// DLL overrides decide if DXVK is actually used or not.
	if !b.cfg.Studio.Dxvk ||
		b.cfg.Studio.DxvkVersion == b.state.Studio.DxvkVersion {
		return nil
	}

	name := filepath.Join(dirs.Cache, "dxvk-"+b.cfg.Studio.DxvkVersion+".tar.gz")
	if _, err := os.Stat(name); err == nil {
		goto install
	}

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return fmt.Errorf("prepare cache: %w", err)
	}

	b.message("Downloading DXVK", "ver", b.cfg.Studio.DxvkVersion)

	if err := netutil.DownloadProgress(
		dxvk.URL(b.cfg.Studio.DxvkVersion), name, &b.pbar); err != nil {
		return fmt.Errorf("download: %w", err)
	}

install:
	defer b.performing()()

	b.message("Extracting DXVK", "ver", b.cfg.Studio.DxvkVersion)

	if err := dxvk.Extract(b.pfx, name); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	b.state.Studio.DxvkVersion = b.cfg.Studio.DxvkVersion
	return nil
}

func (b *bootstrapper) webviewInstaller() string {
	if b.cfg.Studio.WebView == "" {
		return ""
	}
	return filepath.Join(dirs.Cache, "webview-"+b.cfg.Studio.WebView+".exe")
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

	b.message("Fetching WebView", "upload", b.cfg.Studio.WebView)
	d, err := webview.GetDownload(b.cfg.Studio.WebView)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	// TODO: pipe straight to Extract
	tmp, err := os.CreateTemp("", "unc_msedgestandalone.*.exe")
	if err != nil {
		return fmt.Errorf("temp: %w", err)
	}
	defer os.Remove(tmp.Name())

	stop()
	b.message("Downloading WebView", "version", d.Version)
	err = netutil.DownloadProgress(d.URL, tmp.Name(), &b.pbar)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("installer open: %w", err)
	}
	defer f.Close()

	return d.Extract(tmp, f)
}

func (b *bootstrapper) stepWebviewInstall() error {
	name := b.webviewInstaller()

	path := filepath.Join(b.pfx.Dir(), "drive_c/Program Files (x86)/Microsoft/EdgeWebView")
	_, err := os.Stat(path)

	if name == "" && err == nil {
		b.message("Uninstalling WebView")
		os.RemoveAll(path)
		return nil
	} else if err == nil {
		// Already installed
		return nil
	}

	b.message("Installing WebView", "path", name)

	if err := b.pfx.RegistryAdd(`HKCU\Software\Wine\AppDefaults\msedgewebview2.exe`, "Version", "win7"); err != nil {
		return fmt.Errorf("version set: %w", err)
	}

	return run(webview.Install(b.pfx, name))
}
