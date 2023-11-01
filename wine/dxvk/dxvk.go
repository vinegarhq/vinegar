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

const Repo = "https://github.com/doitsujin/dxvk"

func Setenv() {
	log.Printf("Enabling WINE DXVK DLL overrides")

	os.Setenv("WINEDLLOVERRIDES", os.Getenv("WINEDLLOVERRIDES")+";d3d10core=n;d3d11=n;d3d9=n;dxgi=n")
}

func Fetch(name string, ver string) error {
	url := fmt.Sprintf("%s/releases/download/v%[2]s/dxvk-%[2]s.tar.gz", Repo, ver)

	log.Printf("Downloading DXVK %s (%s as %s)", ver, url, name)

	return util.Download(url, name)
}

func Remove(pfx *wine.Prefix) error {
	log.Println("Deleting all overridden DXVK DLLs")

	for _, dir := range []string{"syswow64", "system32"} {
		for _, dll := range []string{"d3d9", "d3d10core", "d3d11", "dxgi"} {
			p := filepath.Join(pfx.Dir(), "drive_c", "windows", dir, dll+".dll")

			log.Println("Removing DXVK DLL:", p)

			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}

	log.Println("Restoring Wineprefix DLLs")

	return pfx.Wine("wineboot", "-u").Run()
}

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
			"x64": filepath.Join(pfx.Dir(), "drive_c", "windows", "system32"),
			"x32": filepath.Join(pfx.Dir(), "drive_c", "windows", "syswow64"),
		}[filepath.Base(filepath.Dir(hdr.Name))]

		if !ok {
			log.Printf("Skipping DXVK unhandled file: %s", hdr.Name)
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

		log.Printf("Extracting DLL %s", p)

		if _, err = io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}

		f.Close()
	}

	log.Printf("Deleting DXVK tarball (%s)", name)
	return os.RemoveAll(name)
}
