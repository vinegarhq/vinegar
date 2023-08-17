package dxvk

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/vinegarhq/aubun/util"
	"github.com/vinegarhq/aubun/wine"
)

const (
	Repo        = "https://github.com/doitsujin/dxvk"
	Version     = "2.2"
	TarName     = "dxvk-" + Version + ".tar.gz"
	WineVarName = "WINEDLLOVERRIDES"
)

var URL = Repo + "/releases/download/v" + Version + "/" + TarName

func Setenv() {
	log.Printf("Using DXVK %s", Version)

	os.Setenv(WineVarName, os.Getenv(WineVarName)+"d3d10core=n;d3d11=n;d3d9=n;dxgi=n")
}

func Fetch(dir string) error {
	tarPath := filepath.Join(dir, TarName)

	if _, err := os.Stat(tarPath); errors.Is(err, os.ErrNotExist) {
		log.Printf("Downloading DXVK %s", Version)

		if err := util.Download(URL, tarPath); err != nil {
			return fmt.Errorf("failed to download DXVK %s: %w", Version, err)
		}
	} else if err == nil {
		log.Printf("DXVK %s is already downloaded", Version)
	} else {
		return err
	}

	return nil
}

func Remove(pfx *wine.Prefix) error {
	log.Println("Removing all overridden DXVK DLLs")

	for _, dir := range []string{"syswow64", "system32"} {
		for _, dll := range []string{"d3d9", "d3d10core", "d3d11", "dxgi"} {
			dllPath := filepath.Join("windows", dir, dll+".dll")

			log.Println("Removing DLL:", dllPath)

			if err := os.Remove(filepath.Join(pfx.Dir, dllPath)); err != nil {
				return err
			}
		}
	}

	log.Println("Restoring wineprefix DLLs")

	return pfx.Exec("wineboot", "-u")
}

func Extract(dir string, pfx *wine.Prefix) error {
	tarPath := filepath.Join(dir, TarName)

	log.Printf("Extracting DXVK %s", Version)

	tarFile, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	stream, err := gzip.NewReader(tarFile)
	if err != nil {
		return err
	}
	defer stream.Close()

	reader := tar.NewReader(stream)

	for {
		header, err := reader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		destDir := map[string]string{
			"x64": filepath.Join(pfx.Dir, "drive_c", "windows", "system32"),
			"x32": filepath.Join(pfx.Dir, "drive_c", "windows", "syswow64"),
		}[filepath.Base(filepath.Dir(header.Name))]

		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return err
		}

		file, err := os.Create(filepath.Join(destDir, path.Base(header.Name)))
		if err != nil {
			return err
		}

		log.Printf("Extracting DLL %s", header.Name)

		if _, err = io.Copy(file, reader); err != nil {
			file.Close()
			return err
		}

		file.Close()
	}

	return nil
}
