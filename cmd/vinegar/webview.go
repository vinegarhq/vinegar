package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/folbricht/pefile" // Cheers to a 5 year old library!
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
)

const (
	WebViewInstallerURL    = "https://catalog.s.download.windowsupdate.com/c/msdownload/update/software/updt/2023/09/microsoftedgestandaloneinstallerx64_1c890b4b8dd6b7c93da98ebdc08ecdc5e30e50cb.exe"
	WebViewTargetInstaller = "MicrosoftEdge_X64_109.0.1518.140.exe.{0D50BFEC-CD6A-4F9A-964C-C7416E3ACB10}"
)

var WebViewInstallerPath = filepath.Join(dirs.Cache, "MicrosoftEdge_X64_109.0.1518.140.exe")

func (b *Binary) InstallWebView() error {
	// This is required for the installer to do some magic
	// that makes it work.
	slog.Info("Setting Wineprefix version to win7")
	b.Splash.SetMessage("Setting up wineprefix")
	if err := b.Prefix.Wine("winecfg", "/v", "win7").Run(); err != nil {
		return err
	}

	b.Splash.SetDesc("109.0.1518.140")

	if _, err := os.Stat(WebViewInstallerPath); err != nil {
		if err := b.DownloadWebView(); err != nil {
			return err
		}
	} else if err == nil {
		slog.Info("WebView installer cached, skipping download", "path", WebViewInstallerPath)
	}

	b.Splash.SetMessage("Installing WebView")
	b.Splash.SetProgress(1.0)
	slog.Info("Running WebView installer", "path", WebViewInstallerPath)

	return b.Prefix.Wine(WebViewInstallerPath,
		"--msedgewebview", "--do-not-launch-msedge", "--system-level",
	).Run()
}

func (b *Binary) DownloadWebView() error {
	b.Splash.SetMessage("Downloading WebView")

	tmp, err := os.CreateTemp("", "unc_msedgestandalone.*.exe")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	slog.Info("Downloading WebView",
		"version", "109.0.1518.140", "url", WebViewInstallerURL, "path", tmp.Name())

	err = netutil.DownloadProgress(WebViewInstallerURL, tmp.Name(), b.Splash.SetProgress)
	if err != nil {
		return err
	}

	b.Splash.SetMessage("Extracting WebView")
	return GetWebViewInstaller(tmp)
}

func GetWebViewInstaller(r io.ReaderAt) error {
	slog.Info("Loading PE file resources")

	inst, err := pefile.New(r)
	if err != nil {
		return err
	}
	defer inst.Close()

	rs, err := inst.GetResources()
	if err != nil {
		return err
	}

	for _, r := range rs {
		if r.Name != "D/102/0" {
			continue
		}

		return ExtractWebView(&r)
	}

	return errors.New("webview installer resource not found")
}

func ExtractWebView(rsrc *pefile.Resource) error {
	slog.Info("Extracting WebView installer", "resource", rsrc.Name)

	r := bytes.NewReader(rsrc.Data)
	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if hdr.Name != WebViewTargetInstaller {
			continue
		}

		exe, err := os.OpenFile(WebViewInstallerPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer exe.Close()

		slog.Info("Extracting WebView installer", "exe", hdr.Name, "path", exe.Name())

		if _, err := io.Copy(exe, tr); err != nil {
			return err
		}

		return nil
	}

	return errors.New("webview installer target not found")
}
