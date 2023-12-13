// Package dxvk implements routines to install DXVK to a given [wine.Prefix]
package dxvk

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/vinegarhq/vinegar/util"
	"github.com/vinegarhq/vinegar/wine"
)

var pfxDll32Path = filepath.Join("drive_c", "Program Files (x86)", "DXVK")
var pfxDll64Path = filepath.Join("drive_c", "Program Files", "DXVK")

// Setenv sets/appends WINEDLLPATH and WINEDLLOVERRIDES to tell Wine
// to use the DXVK DLLs
func Setenv(pfx *wine.Prefix) {
	log.Printf("Enabling WINE DXVK DLL overrides")

	os.Setenv("WINEDLLPATH", os.Getenv("WINEDLLPATH")+
		filepath.Join(pfx.Dir(), pfxDll32Path)+":"+filepath.Join(pfx.Dir(), pfxDll64Path),
	)
	os.Setenv("WINEDLLOVERRIDES", os.Getenv("WINEDLLOVERRIDES")+";d3d10core,d3d11,d3d9,dxgi=n")
}

// Remove will remove the directories the DXVK DLLs have been extracted to by Extract()
func Remove(pfx *wine.Prefix) error {
	log.Println("Removing DXVK DLLs")

	if err := os.RemoveAll(filepath.Join(pfx.Dir(), pfxDll32Path)); err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(pfx.Dir(), pfxDll64Path))
}

// Install will download the DXVK tarball with the given version to a temporary
// file dictated by os.CreateTemp. Afterwards, it will proceed by calling Extract
// with the DXVK tarball, and then removing it.
func Install(ver string, pfx *wine.Prefix) error {
	url := fmt.Sprintf(
		"%s/releases/download/v%[2]s/dxvk-%[2]s.tar.gz",
		"https://github.com/doitsujin/dxvk", ver,
	)

	f, err := os.CreateTemp("", "dxvktarball.*.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	log.Println("Downloading DXVK", ver)
	if err := util.Download(url, f.Name()); err != nil {
		return fmt.Errorf("download dxvk %s: %w", ver, err)
	}

	return Extract(f.Name(), pfx)
}

// Extract extracts the given DXVK tarball named file
// to a folder within the wineprefix.
func Extract(name string, pfx *wine.Prefix) error {
	log.Printf("Extracting DXVK (%s)", name)

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
			"x64": filepath.Join(pfx.Dir(), pfxDll64Path),
			"x32": filepath.Join(pfx.Dir(), pfxDll32Path),
		}[filepath.Base(filepath.Dir(hdr.Name))]

		if !ok {
			return fmt.Errorf("unhandled dxvk tarball file: %s", hdr.Name)
		}

		p := filepath.Join(dir, path.Base(hdr.Name))

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		f, err := os.Create(p)
		if err != nil {
			return err
		}

		log.Printf("Extracting DLL %s", p)

		if _, err = io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}

		f.Close()
	}

	return nil
}
