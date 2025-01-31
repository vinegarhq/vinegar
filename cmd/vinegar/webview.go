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
	WebViewVersion         = "132.0.2957.127"
	WebViewInstallerURL    = "https://catalog.s.download.windowsupdate.com/c/msdownload/update/software/updt/2025/01/microsoftedgestandaloneinstallerx64_8677c34ceecedbf74cc7e4739bf1af4ab1e884d4.exe"
	WebViewInstallerExe    = "MicrosoftEdge_X64_"+WebViewVersion+".exe"
	WebViewInstallerTarget = WebViewInstallerExe + ".{0D50BFEC-CD6A-4F9A-964C-C7416E3ACB10}"
)

var WebViewInstallerPath = filepath.Join(dirs.Cache, WebViewInstallerExe)

func (b *bootstrapper) InstallWebView() error {
	if _, err := os.Stat(WebViewInstallerPath); err != nil {
		if err := b.DownloadWebViewPackage(); err != nil {
			return err
		}
	} else if err == nil {
		slog.Info("WebView installer cached, skipping download", "path", WebViewInstallerPath)
	}

	defer b.Performing()()
	b.status.SetLabel("Installing WebView")
	slog.Info("Running WebView installer", "path", WebViewInstallerPath)

	return b.pfx.Wine(WebViewInstallerPath,
		"--msedgewebview", "--do-not-launch-msedge", "--system-level",
	).Run()
}

func (b *bootstrapper) DownloadWebViewPackage() error {
	b.Message("Downloading WebView Package",
		"version", WebViewVersion, "url", WebViewInstallerURL)

	tmp, err := os.CreateTemp("", "unc_msedgestandalone.*.exe")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	err = netutil.DownloadProgress(WebViewInstallerURL, tmp.Name(), &b.pbar)
	if err != nil {
		return err
	}


	return b.GetWebViewInstaller(tmp)
}

func (b *bootstrapper) GetWebViewInstaller(r io.ReaderAt) error {
	b.Message("Loading WebView package")

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

		return b.ExtractWebView(&r)
	}

	return errors.New("webview installer resource not found")
}

func (b *bootstrapper) ExtractWebView(rsrc *pefile.Resource) error {
	b.Message("Extracting WebView")

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

		if hdr.Name != WebViewInstallerTarget {
			continue
		}

		cac, err := os.OpenFile(WebViewInstallerPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer cac.Close()

		if _, err := io.Copy(cac, tr); err != nil {
			return err
		}

		return nil
	}

	return errors.New("webview installer target not found")
}
