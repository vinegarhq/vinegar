package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Package struct {
	Name       string
	Signature  string
	PackedSize int64
	Size       int64
}

type PackageManifest struct {
	Version  string
	Packages []Package
}

func GetLatestVersion(versionFile string) string {
	version, err := GetURLBody("https://setup.rbxcdn.com/" + versionFile)

	if err != nil {
		log.Fatal(err)
	}

	return version
}

func (m *PackageManifest) Construct() {
	log.Println("Constructing Package Manifest")

	if len(m.Version) != 24 {
		log.Fatal("invalid version set")
	}

	rawManifest, err := GetURLBody("https://setup.rbxcdn.com/" + m.Version + "-rbxPkgManifest.txt")

	if err != nil {
		log.Fatal(err)
	}

	manif := strings.Split(rawManifest, "\r\n")
	index := 1

	if manif[0] != "v0" {
		log.Fatal("invalid package manifest version", manif[0])
	}

	for {
		if index+4 > len(manif) {
			break
		}

		fileName := manif[index]
		signature := manif[index+1]
		packedSize, _ := strconv.ParseInt(manif[index+2], 10, 64)
		size, _ := strconv.ParseInt(manif[index+3], 10, 64)
		index += 4

		if fileName == "RobloxPlayerLauncher.exe" {
			continue
		}

		m.Packages = append(m.Packages, Package{
			Name:       fileName,
			Signature:  signature,
			PackedSize: packedSize,
			Size:       size,
		})
	}
}

func (p *Package) Download(version string) {
	packageURL := "https://setup.rbxcdn.com/" + version + "-" + p.Name
	packagePath := filepath.Join(Dirs.Downloads, p.Signature)

	if _, err := os.Stat(packagePath); err == nil {
		log.Println("Found", packagePath)

		return
	}

	if err := Download(packageURL, packagePath); err != nil {
		log.Fatalf("failed to download package %s: %s", p.Name, err)
	}
}

func (m *PackageManifest) DownloadVerifyExtractAll(directories map[string]string) {
	var waitGroup sync.WaitGroup

	waitGroup.Add(len(m.Packages))

	for _, pkg := range m.Packages {
		go func(ver string, pkg Package) {
			pkg.Download(ver)
			waitGroup.Done()
		}(m.Version, pkg)
	}

	waitGroup.Wait()

	for _, pkg := range m.Packages {
		VerifyFileMD5(filepath.Join(Dirs.Downloads, pkg.Signature), pkg.Signature)
	}

	waitGroup.Add(len(m.Packages))

	for _, pkg := range m.Packages {
		go func(ver string, pkg Package, dirs map[string]string) {
			packagePath := filepath.Join(Dirs.Downloads, pkg.Signature)
			packageDirDest := filepath.Join(LocalProgramDir, ver, dirs[pkg.Name])

			CreateDirs(packageDirDest)

			if err := UnzipFolder(packagePath, packageDirDest); err != nil {
				log.Fatalf("failed to extract package %s: %s", pkg.Name, err)
			}

			waitGroup.Done()
		}(m.Version, pkg, directories)
	}

	waitGroup.Wait()
}
