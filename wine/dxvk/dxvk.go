// Package dxvk implements routines to install DXVK to a given [wine.Prefix]
package dxvk

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/vinegarhq/vinegar/internal/netutil"
	"github.com/vinegarhq/vinegar/wine"
)

const Repo = "https://github.com/doitsujin/dxvk"

// Setenv sets/appends WINEDLLOVERRIDES to tell Wine to use the DXVK DLLs.
//
// This is required to call inorder to tell Wine to use DXVK.
func Setenv() {
	slog.Info("Enabling WINE DXVK DLL overrides")

	os.Setenv("WINEDLLOVERRIDES", os.Getenv("WINEDLLOVERRIDES")+";d3d10core=n;d3d11=n;d3d9=n;dxgi=n")
}

func Remove(pfx *wine.Prefix) error {
	slog.Info("Deleting DXVK DLLs", "pfx", pfx)

	for _, dir := range []string{"syswow64", "system32"} {
		for _, dll := range []string{"d3d9", "d3d10core", "d3d11", "dxgi"} {
			p := filepath.Join(pfx.Dir(), "drive_c", "windows", dir, dll+".dll")

			slog.Info("Removing DXVK overriden Wine DLL", "path", p)

			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}

	slog.Info("Restoring Wineprefix DLLs", "pfx", pfx)

	return pfx.Wine("wineboot", "-u").Run()
}

// Install will download the DXVK tarball with the given version to a temporary
// file dictated by os.CreateTemp. Afterwards, it will proceed by calling Extract
// with the DXVK tarball, and then removing it.
func Install(ver string, pfx *wine.Prefix) error {
	url := fmt.Sprintf("%s/releases/download/v%[2]s/dxvk-%[2]s.tar.gz", Repo, ver)
	f, err := os.CreateTemp("", "dxvktarball.*.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	slog.Info("Downloading DXVK tarball", "url", url, "path", f.Name())

	if err := netutil.Download(url, f.Name()); err != nil {
		return fmt.Errorf("download dxvk %s: %w", ver, err)
	}

	if err := Extract(f.Name(), pfx); err != nil {
		return fmt.Errorf("extract dxvk %s: %w", ver, err)
	}

	return nil
}

func Extract(name string, pfx *wine.Prefix) error {
	slog.Info("Extracting DXVK", "file", name, "pfx", pfx)

	tf, err := os.Open(name)
	if err != nil {
		return err
	}
	defer tf.Close()

	zr, err := gzip.NewReader(tf)
	if err != nil {
		return err
	}
	defer zr.Close()

	tr := tar.NewReader(zr)

	for {
		hdr, err := tr.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		dir, ok := map[string]string{
			"x64": filepath.Join(pfx.Dir(), "drive_c", "windows", "system32"),
			"x32": filepath.Join(pfx.Dir(), "drive_c", "windows", "syswow64"),
		}[filepath.Base(filepath.Dir(hdr.Name))]

		if !ok {
			slog.Warn("Skipping DXVK unhandled file", "file", hdr.Name)
			continue
		}

		p := filepath.Join(dir, path.Base(hdr.Name))

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		f, err := os.Create(p)
		if err != nil {
			return err
		}

		slog.Info("Extracting DXVK DLL", "path", p)

		if _, err = io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}

		f.Close()
	}

	slog.Info("Deleting DXVK tarball", "path", name)
	return os.RemoveAll(name)
}
