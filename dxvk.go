package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

const (
	DXVKVER = "2.1"
	DXVKTAR = "dxvk-" + DXVKVER + ".tar.gz"
	DXVKURL = "https://github.com/doitsujin/dxvk/releases/download/v" + DXVKVER + "/" + DXVKTAR
)

// DxvkInstallMarker file, created on DXVK installation and removed at DXVK
// uninstallation, is an easy way to tell if DXVK is installed, necessary for
// automatic installation and uninstallation of DXVK seen in DxvkStrap().
var DxvkInstallMarker = filepath.Join(Dirs.Pfx, ".vinegar-dxvk")

func DxvkMarkerExist() bool {
	_, err := os.Open(DxvkInstallMarker)

	return err == nil
}

func DxvkStrap() {
	if Config.Dxvk {
		DxvkInstall(false)

		Config.Renderer = "D3D11"
		Config.Env["WINEDLLOVERRIDES"] += "d3d10core=n;d3d11=n;d3d9=n;dxgi=n"
		os.Setenv("WINEDLLOVERRIDES", Config.Env["WINEDLLOVERRIDES"])

		return
	}

	DxvkUninstall(false)
}

func DxvkExtract(tarball string) {
	var winDir string

	log.Println("Extracting DXVK")

	dxvkTarball, err := os.Open(tarball)

	if err != nil {
		log.Fatal(err)
	}

	dxvkGzip, err := gzip.NewReader(dxvkTarball)
	if err != nil {
		log.Fatal(err)
	}

	dxvkTar := tar.NewReader(dxvkGzip)

	for header, err := dxvkTar.Next(); err == nil; header, err = dxvkTar.Next() {
		if header.Typeflag != tar.TypeReg {
			continue
		}

		switch path.Base(path.Dir(header.Name)) {
		case "x64":
			winDir = "system32"
		case "x32":
			winDir = "syswow64"
		default:
			continue
		}

		winDir := filepath.Join(Dirs.Pfx, "drive_c", "windows", winDir)

		// just in case
		CreateDirs(winDir)

		writer, err := os.Create(filepath.Join(winDir, path.Base(header.Name)))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Extracting DLL:", writer.Name())

		if _, err = io.Copy(writer, dxvkTar); err != nil {
			log.Fatal(err)
		}
	}
}

func DxvkInstall(force bool) {
	if !force {
		if DxvkMarkerExist() {
			return
		}
	}

	dxvkTarballPath := filepath.Join(Dirs.Cache, DXVKTAR)

	if _, err := os.Stat(dxvkTarballPath); errors.Is(err, os.ErrNotExist) {
		if err := Download(DXVKURL, dxvkTarballPath); err != nil {
			log.Fatal(err)
		}
	}

	DxvkExtract(dxvkTarballPath)

	if _, err := os.Create(DxvkInstallMarker); err != nil {
		log.Fatal(err)
	}
}

func DxvkUninstall(force bool) {
	if !force {
		if !DxvkMarkerExist() {
			return
		}
	}

	log.Println("Uninstalling DXVK")

	for _, dir := range []string{"syswow64", "system32"} {
		for _, dll := range []string{"d3d9", "d3d10core", "d3d11", "dxgi"} {
			dllFile := filepath.Join(Dirs.Pfx, "drive_c", "windows", dir, dll+".dll")
			log.Println("Removing DLL:", dllFile)

			if err := os.RemoveAll(dllFile); err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Println("Updating wineprefix")

	// Updating the wineprefix is necessary, since the DLLs
	// that were overrided by DXVK, were subsequently deleted,
	// and has to be restored (updated)
	if err := Exec("wineboot", "", "-u"); err != nil {
		log.Fatal(err)
	}

	if err := os.RemoveAll(DxvkInstallMarker); err != nil {
		log.Fatal(err)
	}
}
