package main

import (
	"errors"
	"fmt"
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

	if !b.pfx.Exists() {
		if err := b.stepPrefixInit(); err != nil {
			return fmt.Errorf("init: %w", err)
		}
	}

	if err := b.checkPrefix(); err != nil {
		return err
	}

	if err := b.stepDxvkInstall(); err != nil {
		return fmt.Errorf("dxvk: %w", err)
	}

	if err := b.setupWebView(); err != nil {
		return fmt.Errorf("webview: %w", err)
	}

	return nil
}

func (b *bootstrapper) stepPrefixInit() error {
	defer b.performing()()

	b.message("Initializing Wineprefix", "dir", b.pfx.Dir())
	return run(b.pfx.Init())
}

func (b *bootstrapper) checkPrefix() error {
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

func (b *bootstrapper) stepDxvkInstall() error {
	if !b.cfg.Studio.Dxvk ||
		b.cfg.Studio.DxvkVersion == b.state.Studio.DxvkVersion {
		return nil
	}

	ver := b.cfg.Studio.DxvkVersion
	dxvkPath := filepath.Join(dirs.Cache, "dxvk-"+ver+".tar.gz")

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return err
	}

	if _, err := os.Stat(dxvkPath); err != nil {
		url := dxvk.URL(ver)

		b.message("Downloading DXVK", "ver", ver)

		if err := netutil.DownloadProgress(url, dxvkPath, &b.pbar); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	defer b.performing()()

	b.message("Extracting DXVK", "version", ver)

	if err := dxvk.Extract(b.pfx, dxvkPath); err != nil {
		return err
	}

	b.state.Studio.DxvkVersion = ver

	return nil
}

func (b *bootstrapper) setupWebView() error {
	path := filepath.Join(b.pfx.Dir(), "drive_c/Program Files (x86)/Microsoft/EdgeWebView")
	// Since we do not keep webview tracked in state, simply removing it
	// always if it is disabled is fine, since it will be a harmless error.
	// TODO: implement a webview uninstaller
	if b.cfg.Studio.WebView == "" {
		os.RemoveAll(path)
		return nil
	}

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	return b.stepWebviewInstall()
}

func (b *bootstrapper) stepWebviewInstall() error {
	name := filepath.Join(dirs.Cache, "webview-"+b.cfg.Studio.WebView+".exe")
	if _, err := os.Stat(name); err != nil {
		if err := b.stepWebviewDownload(name); err != nil {
			return err
		}
	}
	defer b.performing()()

	b.message("Installing WebView", "path", name)

	if err := b.pfx.RegistryAdd(`HKCU\Software\Wine\AppDefaults\msedgewebview2.exe`, "Version", "win7"); err != nil {
		return fmt.Errorf("version set: %w", err)
	}

	return run(webview.Install(b.pfx, name))
}

func (b *bootstrapper) stepWebviewDownload(name string) error {
	stop := b.performing()

	b.message("Fetching WebView", "upload", b.cfg.Studio.WebView)
	d, err := webview.GetDownload(b.cfg.Studio.WebView)
	if err != nil {
		return err
	}

	// TODO: pipe straight to Extract
	tmp, err := os.CreateTemp("", "unc_msedgestandalone.*.exe")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	stop()
	b.message("Downloading WebView", "version", d.Version)
	err = netutil.DownloadProgress(d.URL, tmp.Name(), &b.pbar)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	return d.Extract(tmp, f)
}
