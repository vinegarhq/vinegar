package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sewnie/wine"
	"github.com/sewnie/wine/dxvk"
	"github.com/sewnie/wine/peutil"
	"github.com/sewnie/wine/webview2"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"

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

	installed, err := dxvk.DLLVersion(filepath.Join(b.dir, "d3d11.dll"))
	var native string
	if errors.Is(err, os.ErrNotExist) {
		// DXVK is enabled, but is not currently installed for Studio.
		// Check if it is present inside the Wineprefix instead.
		// Either installation solution will work.
		installed, err = dxvk.Version(b.pfx)
		native = installed
	}
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

	if native != "" {
		// If DXVK is installed in the wineprefix, and it was deemed
		// out of date, prefer to restore the original DLLs to prefer
		// the new installation method. but only if installing or
		// updating DXVK anew.
		//
		// Yes, this will show a dialog to the user, but it will be
		// so very brief, and the only other solution is to restart
		// the entire wineserver.
		//
		// It is also optional, so ignoring the error is preferred,
		// as Wine will prefer the DLLs in the new installation method.
		slog.Info("Restoring DirectX DLLs")
		_ = dxvk.Restore(b.pfx)
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	b.message(L("Extracting DXVK"), "version", version)

	zr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer zr.Close()

	tr := tar.NewReader(zr)

	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		if filepath.Ext(hdr.Name) != ".dll" {
			continue
		}

		if filepath.Base(filepath.Dir(hdr.Name)) != "x64" {
			slog.Debug("Ignoring DXVK installation entry", "file", hdr.Name)
			continue
		}

		name := filepath.Join(b.dir, filepath.Base(hdr.Name))
		slog.Debug("Installing DXVK DLL locally", "dest", name)

		f, err := os.Create(name)
		if err != nil {
			return err
		}

		if _, err = io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
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
