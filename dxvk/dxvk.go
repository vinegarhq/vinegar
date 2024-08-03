// Package dxvk implements routines to install DXVK to a given [wine.Prefix]
package dxvk

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/apprehensions/wine"
)

const Repo = "https://github.com/doitsujin/dxvk"

// Setenv sets/appends WINEDLLOVERRIDES to tell Wine to use the DXVK DLLs.
//
// This is required to call inorder to tell Wine to use it's own DLLs,
// which is presumably overwritten by Extract.
func Setenv() {
	slog.Info("Enabling WINE DXVK DLL overrides")

	os.Setenv("WINEDLLOVERRIDES", os.Getenv("WINEDLLOVERRIDES")+";d3d9,d3d10core,d3d11,dxgi=n")
}

// Remove removes the DXVK overridden DLLs from the
// given wineprefix, then restores global wine DLLs.
func Remove(pfx *wine.Prefix) error {
	slog.Info("Deleting DXVK DLLs", "pfx", pfx)

	for _, dir := range []string{"syswow64", "system32"} {
		for _, dll := range []string{"d3d9", "d3d10core", "d3d11", "dxgi"} {
			p := filepath.Join(pfx.Dir(), "drive_c", "windows", dir, dll+".dll")

			slog.Info("Removing DXVK overriden Wine DLL", "path", p)

                        // Checks to see if the given file exists, if not, skips it and continues removal process
                        if _, err := os.Stat(p); os.IsNotExist(err) {
                                slog.Info("File does not exist, skipping", "path", p)
                                continue
                        }

			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}

	slog.Info("Restoring Wineprefix DLLs", "pfx", pfx)

	return pfx.Wine("wineboot", "-u").Run()
}

// URL returns the DXVK tarball URL for the given version.
func URL(ver string) string {
	return fmt.Sprintf("%s/releases/download/v%[2]s/dxvk-%[2]s.tar.gz", Repo, ver)
}

// Extract extracts DXVK's DLLs into the given wineprefix -
// overriding Wine's D3D DLLs, given the path to a DXVK tarball.
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

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		// Theoretically this is unsafe becuase there is no standard to DLL
		// tarballs other than it having a directory structure like
		// x64, x32, x86, with each folder containing a set of dlls
		// that can be installed to a folder.
		//
		// Oh well.

		if filepath.Ext(hdr.Name) != ".dll" {
			slog.Warn("Skipping DXVK unhandled file", "file", hdr.Name)
			continue
		}

		dir := "system32"
		if filepath.Base(filepath.Dir(hdr.Name)) == "x32" {
			dir = "syswow64"
		}

		dst := filepath.Join(pfx.Dir(), "drive_c", "windows", dir, filepath.Base(hdr.Name))

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}

		f, err := os.Create(dst)
		if err != nil {
			return err
		}

		slog.Info("Extracting DXVK DLL", "dest", dst)

		if _, err = io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}

		f.Close()
	}

	return nil
}
