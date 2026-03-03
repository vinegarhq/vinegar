package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sewnie/wine"
	"github.com/sewnie/wine/peutil"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/sewnie/wine/dxvk"
	"github.com/vinegarhq/vinegar/internal/netutil"
	"github.com/sewnie/wine/webview2"

	. "github.com/pojntfx/go-gettext/pkg/i18n"
)

func (b *bootstrapper) setupDXVK() error {
	version := b.cfg.Studio.DXVKVersion()
	if version == "" {
		return nil
	}

	// If DXVK is installed in the wineprefix, uninstallation
	// won't be necessary if it's disabled as it still requires
	// DLL overrides to be present.
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

func (b *bootstrapper) webViewInstaller() string {
	// TODO: Clear old downloads here.
	return filepath.Join(dirs.Cache, "webview-"+b.cfg.Studio.WebView+".exe")
}

func (b *bootstrapper) webViewVersion(offline *wine.Registry) string {
	if offline == nil {
		// Wineprefix is not initialized
		return ""
	}

	k := offline.Query(webview2.VersionPath)
	if k == nil {
		// WebView is not installed
		return ""
	}

	if v := k.GetValue("DisplayVersion"); v != nil {
		return v.Data.(string)
	}

	slog.Warn("WebView2 installed status missing, assuming uninstalled")
	return ""
}

// downloadWebView, when WebView is enabled in the Studio configuration,
// will prepare the WebView installer, prior to initializing the wineprefix
// and running the installer.
func (b *bootstrapper) downloadWebView(installed string) error {
	defer b.performing()()
	inst := b.webViewInstaller()
	if installed == b.cfg.Studio.WebView || b.cfg.Studio.WebView == "" {
		return nil
	}

	// Ensures the installer is a valid executable
	f, err := peutil.Open(inst)
	if err == nil {
		return nil
	}
	f.Close()

	b.message(L("Fetching WebView"), "upload", b.cfg.Studio.WebView)

	// Microsoft doesn't like compressed requests
	webview2.Client.Transport.(*http.Transport).DisableCompression = true
	d, err := webview2.Stable.Runtime(b.cfg.Studio.WebView, "x64")
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	b.message(L("Downloading WebView"), "catalog", d.Delivery.CatalogID)
	return netutil.DownloadProgress(d.URL, inst, &b.pbar)
}

// installWebView checks the Studio WebView version and installs WebView
// if the version is out of date or requires installation. The previous
// version of WebView will be uninstalled before any installation.
func (b *bootstrapper) installWebView(installed string) error {
	version := b.cfg.Studio.WebView

	b.message(L("Checking WebView"), "against", version)

	if installed != "" && installed != version {
		b.message(L("Uninstalling WebView"), "current", installed, "new", version)
		if err := webview2.Uninstall(b.pfx, installed); err != nil {
			return fmt.Errorf("uninstall: %w", err)
		}
	}
	if installed == version || version == "" {
		return nil
	}

	inst := b.webViewInstaller()
	b.message(L("Installing WebView"), "version", version, "path", inst)
	defer b.performing()()

	return webview2.Install(b.pfx, inst)
}
