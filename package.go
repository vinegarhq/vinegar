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

func (r *Roblox) GetPackages() {
	log.Println("Constructing Package Manifest")

	if len(r.Version) != 24 {
		log.Fatal("invalid version set")
	}

	rawManifest, err := GetURLBody(r.URL + r.Version + "-rbxPkgManifest.txt")
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

		r.Packages = append(r.Packages, Package{
			Name:       fileName,
			Signature:  signature,
			PackedSize: packedSize,
			Size:       size,
		})
	}
}

func (r *Roblox) DownloadVerifyExtractAll() {
	var waitGroup sync.WaitGroup

	waitGroup.Add(len(r.Packages))

	for _, pkg := range r.Packages {
		go func(url string, ver string, pkg Package) {
			packageURL := url + ver + "-" + pkg.Name
			packagePath := filepath.Join(Dirs.Downloads, pkg.Signature)

			if _, err := os.Stat(packagePath); err == nil {
				log.Println("Found", packagePath)

				return
			}

			if err := Download(packageURL, packagePath); err != nil {
				log.Fatalf("failed to download package %s: %s", pkg.Name, err)
			}

			waitGroup.Done()
		}(r.URL, r.Version, pkg)
	}

	waitGroup.Wait()

	for _, pkg := range r.Packages {
		VerifyFileMD5(filepath.Join(Dirs.Downloads, pkg.Signature), pkg.Signature)
	}

	waitGroup.Add(len(r.Packages))
	CreateDirs(r.VersionDir)

	for _, pkg := range r.Packages {
		go func(pkg Package, dirs map[string]string) {
			packagePath := filepath.Join(Dirs.Downloads, pkg.Signature)
			packageDirDest := filepath.Join(r.VersionDir, dirs[pkg.Name])

			if err := UnzipFolder(packagePath, packageDirDest); err != nil {
				log.Fatalf("failed to extract package %s: %s", pkg.Name, err)
			}

			waitGroup.Done()
		}(pkg, r.Directories)
	}

	waitGroup.Wait()
}
