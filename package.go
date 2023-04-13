package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
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

func GetLatestVersion() string {
	version, err := GetUrlBody("https://setup.rbxcdn.com/version")

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

	rawManifest, err := GetUrlBody("https://setup.rbxcdn.com/" + m.Version + "-rbxPkgManifest.txt")

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

func (m *PackageManifest) DownloadAll() {
	var wg sync.WaitGroup

	wg.Add(len(m.Packages))
	for _, pkg := range m.Packages {

		go func(ver string, pkg Package) {
			packageUrl := "https://setup.rbxcdn.com/" + ver + "-" + pkg.Name
			packagePath := filepath.Join(Dirs.Downloads, pkg.Signature)

			if _, err := os.Stat(packagePath); err == nil {
				log.Println("Found", packagePath)
				wg.Done()
				return
			}

			if err := Download(packageUrl, packagePath); err != nil {
				log.Fatalf("failed to download package %s: %s", pkg.Name, err)
			}
			wg.Done()
		}(m.Version, pkg)
	}

	wg.Wait()
}

func (m *PackageManifest) VerifyAll() {
	for _, pkg := range m.Packages {

		log.Printf("Verifying Package %s: %s", pkg.Name, pkg.Signature)
		hash := md5.New()

		packagePath := filepath.Join(Dirs.Downloads, pkg.Signature)
		packageFile, err := os.Open(packagePath)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := io.Copy(hash, packageFile); err != nil {
			log.Fatal(err)
		}

		if pkg.Signature != hex.EncodeToString(hash.Sum(nil)) {
			log.Fatalf("Package %s checksum mismatch: %x", pkg.Name, hash.Sum(nil))
		}

		packageFile.Close()
	}
}

func (m *PackageManifest) ExtractAll(directories map[string]string) {
	for _, pkg := range m.Packages {
		packagePath := filepath.Join(Dirs.Downloads, pkg.Signature)
		packageDirDest := filepath.Join(Dirs.Versions, m.Version, directories[pkg.Name])

		CreateDirs(packageDirDest)
		if err := UnzipFolder(packagePath, packageDirDest); err != nil {
			log.Fatalf("failed to extract package %s: %s", pkg.Name, err)
		}
	}
}

func ClientPackageDirectories() map[string]string {
	return map[string]string{
		"RobloxApp.zip":                 "",
		"shaders.zip":                   "shaders",
		"ssl.zip":                       "ssl",
		"content-avatar.zip":            "content/avatar",
		"content-configs.zip":           "content/configs",
		"content-fonts.zip":             "content/fonts",
		"content-sky.zip":               "content/sky",
		"content-sounds.zip":            "content/sounds",
		"content-textures2.zip":         "content/textures",
		"content-models.zip":            "content/models",
		"content-textures3.zip":         "PlatformContent/pc/textures",
		"content-terrain.zip":           "PlatformContent/pc/terrain",
		"content-platform-fonts.zip":    "PlatformContent/pc/fonts",
		"extracontent-luapackages.zip":  "ExtraContent/LuaPackages",
		"extracontent-translations.zip": "ExtraContent/translations",
		"extracontent-models.zip":       "ExtraContent/models",
		"extracontent-textures.zip":     "ExtraContent/textures",
		"extracontent-places.zip":       "ExtraContent/places",
	}
}
