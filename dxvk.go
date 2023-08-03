package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/vinegarhq/vinegar/util"
)

// this file, created on DXVK installation and removed at DXVK uninstallation,
// is an easy way to tell if DXVK is installed, necessary for automatic installation
// and uninstallation of DXVK.
var DxvkInstallMarker = filepath.Join(Dirs.Prefix, ".vinegar-dxvk")

const DXVKVER = "2.2"

func DxvkStrap() {
	_, err := os.Stat(DxvkInstallMarker)

	// Uninstall DXVK when DXVK is disabled, and if the file exists (no error)
	if err == nil {
		if !Config.Dxvk {
			DxvkUninstall()
		}

		return
	}

	DxvkInstall()
}

func DxvkInstall() {
	log.Printf("Installing DXVK %s", DXVKVER)

	tarName := "dxvk-" + DXVKVER + ".tar.gz"
	tarPath := filepath.Join(Dirs.Cache, tarName)
	url := "https://github.com/doitsujin/dxvk/releases/download/v" + DXVKVER + "/" + tarName

	// Check if the DXVK tarball exists, if not; download it.
	// Catch any other errors by stat(2)
	if _, err := os.Stat(tarPath); errors.Is(err, os.ErrNotExist) {
		if err := util.Download(url, tarPath); err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	if err := DxvkExtract(tarPath); err != nil {
		log.Fatal(err)
	}

	// After installation, create the marker file
	if _, err := os.Create(DxvkInstallMarker); err != nil {
		log.Fatal(err)
	}
}

// Extracting the decompression to another function in the util library is
// bigger than the final used function, and it will be unrealistic, since
// all we want in the tarball is the specific set of DLLs, but not to extract
// them to a specific directory, which requires scanning it, and copying the
// required DLLs, which in return is bigger than the function below.
func DxvkExtract(source string) error {
	log.Println("Extracting DXVK")

	tarFile, err := os.Open(source)
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

		// Only need the DLL files
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Uses the DLL's parent directory {x64, x32) to determine where
		// it's installation directory is set to. map["x64"] -> system32
		destDir := map[string]string{
			"x64": filepath.Join(Dirs.Prefix, "drive_c", "windows", "system32"),
			"x32": filepath.Join(Dirs.Prefix, "drive_c", "windows", "syswow64"),
		}[filepath.Base(filepath.Dir(header.Name))]

		if err := os.MkdirAll(destDir, DirMode); err != nil {
			return err
		}

		file, err := os.Create(filepath.Join(destDir, path.Base(header.Name)))
		if err != nil {
			return err
		}

		log.Println("Extracting DLL:", file.Name())

		if _, err = io.Copy(file, reader); err != nil {
			file.Close()
			return err
		}

		file.Close()
	}

	return nil
}

func DxvkUninstall() {
	log.Println("Uninstalling DXVK")

	for _, dir := range []string{"syswow64", "system32"} {
		for _, dll := range []string{"d3d9", "d3d10core", "d3d11", "dxgi"} {
			dllFile := filepath.Join(Dirs.Prefix, "drive_c", "windows", dir, dll+".dll")
			log.Println("Removing DLL:", dllFile)

			if err := os.Remove(dllFile); err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Println("Updating wineprefix")

	// Updating the wineprefix is necessary, since the DLLs
	// that were overrided by DXVK, were subsequently deleted,
	// and has to be restored (updated)
	if err := exec.Command("wineboot", "-u").Run(); err != nil {
		log.Fatal(err)
	}

	// Remove the file that indicates DXVK installation.
	if err := os.Remove(DxvkInstallMarker); err != nil {
		log.Fatal(err)
	}
}
