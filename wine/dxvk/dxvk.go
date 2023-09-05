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

	"github.com/vinegarhq/vinegar/util"
	"github.com/vinegarhq/vinegar/wine"
)

const Repo = "https://github.com/doitsujin/dxvk"

func Setenv() {
	log.Printf("Enabling DXVK DLL overrides")

	os.Setenv("WINEDLLOVERRIDES", os.Getenv("WINEDLLOVERRIDES")+"d3d10core=n;d3d11=n;d3d9=n;dxgi=n")
}

func Fetch(name string, ver string) error {
	url := fmt.Sprintf("%s/releases/download/v%[2]s/dxvk-%[2]s.tar.gz", Repo, ver)

	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		log.Printf("Downloading DXVK %s", ver)

		if err := util.Download(url, name); err != nil {
			return fmt.Errorf("failed to download DXVK: %w", err)
		}
	} else if err == nil {
		log.Printf("DXVK %s is already downloaded", ver)
	} else {
		return err
	}

	return nil
}

func Remove(pfx *wine.Prefix) error {
	log.Println("Removing all overridden DXVK DLLs")

	for _, dir := range []string{"syswow64", "system32"} {
		for _, dll := range []string{"d3d9", "d3d10core", "d3d11", "dxgi"} {
			dllPath := filepath.Join("drive_c", "windows", dir, dll+".dll")

			log.Println("Removing DLL:", dllPath)

			if err := os.Remove(filepath.Join(pfx.Dir, dllPath)); err != nil {
				return err
			}
		}
	}

	log.Println("Restoring wineprefix DLLs")

	return pfx.Exec("wineboot", "-u")
}

func Extract(name string, pfx *wine.Prefix) error {
	log.Printf("Extracting DXVK")

	tarFile, err := os.Open(name)
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

		destDir, ok := map[string]string{
			"x64": filepath.Join(pfx.Dir, "drive_c", "windows", "system32"),
			"x32": filepath.Join(pfx.Dir, "drive_c", "windows", "syswow64"),
		}[filepath.Base(filepath.Dir(header.Name))]

		if !ok {
			log.Printf("Skipping DXVK unhandled file: %s", header.Name)
			continue
		}

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
