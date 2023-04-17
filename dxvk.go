package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// this file, created on DXVK installation and removed at DXVK uninstallation,
// is an easy way to tell if DXVK is installed, necessary for automatic installation
// and uninstallation of DXVK.
var DxvkInstallMarker = filepath.Join(Dirs.Pfx, ".vinegar-dxvk")

const DXVKVER = "2.1"

func DxvkStrap() error {
	_, err := os.Stat(DxvkInstallMarker)

	// Uninstall DXVK when DXVK is disabled, and if the file exists (no error)
	if !Config.Dxvk && err == nil {
		if err = DxvkUninstall(); err != nil {
			return fmt.Errorf("failed to uninstall dxvk: %w", err)
		}
	}

	// ENOENT is all we care about, any other errors are then catched.
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	log.Printf("Installing DXVK %s", DXVKVER)

	if err := DxvkInstall(); err != nil {
		return fmt.Errorf("failed to install dxvk: %w", err)
	}

	// Forces Roblox to use the D3D11 renderer, which is needed by
	// DXVK to do its thing.
	Config.Renderer = "D3D11"

	// Tells wine to use the DXVK DLLs
	Config.Env["WINEDLLOVERRIDES"] += "d3d10core=n;d3d11=n;d3d9=n;dxgi=n"
	os.Setenv("WINEDLLOVERRIDES", Config.Env["WINEDLLOVERRIDES"])

	return nil
}

func DxvkInstall() error {
	tarName := "dxvk-" + DXVKVER + ".tar.gz"
	tarPath := filepath.Join(Dirs.Cache, tarName)
	url := "https://github.com/doitsujin/dxvk/releases/download/v" + DXVKVER + "/" + tarName

	// Check if the DXVK tarball exists, if not; download it.
	// Catch any other errors by stat(2)
	if _, err := os.Stat(tarPath); errors.Is(err, os.ErrNotExist) {
		if err := Download(url, tarPath); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if err := DxvkExtract(tarPath); err != nil {
		return err
	}

	// After installation, create the marker file
	if _, err := os.Create(DxvkInstallMarker); err != nil {
		return fmt.Errorf("dxvk marker file: %w", err)
	}

	return nil
}

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
			"x64": filepath.Join(Dirs.Pfx, "drive_c", "windows", "system32"),
			"x32": filepath.Join(Dirs.Pfx, "drive_c", "windows", "syswow64"),
		}[filepath.Base(filepath.Dir(header.Name))]

		if err := Mkdirs(destDir); err != nil {
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

func DxvkUninstall() error {
	log.Println("Uninstalling DXVK")

	for _, dir := range []string{"syswow64", "system32"} {
		for _, dll := range []string{"d3d9", "d3d10core", "d3d11", "dxgi"} {
			dllFile := filepath.Join(Dirs.Pfx, "drive_c", "windows", dir, dll+".dll")
			log.Println("Removing DLL:", dllFile)

			if err := os.Remove(dllFile); err != nil {
				return err
			}
		}
	}

	log.Println("Updating wineprefix")

	// Updating the wineprefix is necessary, since the DLLs
	// that were overrided by DXVK, were subsequently deleted,
	// and has to be restored (updated)
	if err := exec.Command("wineboot", "-u").Run(); err != nil {
		return err
	}

	// Remove the file that indicates DXVK installation.
	if err := os.Remove(DxvkInstallMarker); err != nil {
		return fmt.Errorf("dxvk marker file: %w", err)
	}

	return nil
}
