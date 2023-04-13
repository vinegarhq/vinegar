package main

import (
	"log"
	"os"
	"strings"
	"strconv"
	"io"
	"path/filepath"
	"archive/zip"
)

type Package struct {
	Name string
	Signature string
	PackedSize int64
	Size int64
}

func ConstructPackages(rawManifest string) []Package {
	var packages []Package
	log.Println("Constructing Package Manifest")

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

		packages = append(packages, Package{
			Name: fileName,
			Signature: signature,
			PackedSize: packedSize,
			Size: size,
		})
	}

	return packages
}

func DownloadPackage(version string, destDir string, pkg Package) {
	packageUrl := "https://setup.rbxcdn.com/"+version+"-"+pkg.Name
	packagePath := filepath.Join(destDir, pkg.Signature)

	if _, err := os.Stat(packagePath); err == nil {
		return
	}
	
	Download(packageUrl, packagePath)
}

func ExtractPackage(pkg Package, sourceDir string, destDir string) {
	packageDirectories := map[string]string{
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

	packagePath := filepath.Join(sourceDir, pkg.Signature)
	packageDest := filepath.Join(destDir, packageDirectories[pkg.Name])

	UnzipFolder(packagePath, packageDest)
}

func UnzipFolder(source string, destDir string) {
	log.Println("Extracting", source)

	zip, err := zip.OpenReader(source)
	if err != nil {
		panic(err)
	}

	for _, file := range zip.File {
		filePath := filepath.Join(destDir, file.Name)
		log.Println("Unzipping", filePath)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			panic(err)
		}
	
		fileZipped, err := file.Open()
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(destFile, fileZipped); err != nil {
			panic(err)
		}
	
		destFile.Close()
		fileZipped.Close()
	}

	zip.Close()
}

//func main() {
//	log.Println("Today is a good day :D")
//	ver := GetUrlBody("https://setup.rbxcdn.com/version")
//
//	if err := os.MkdirAll(ver, 0755); err != nil {
//		panic(err)
//	}
//
//	log.Println("Version", ver)
//
//	manif := GetUrlBody("https://setup.rbxcdn.com/" + ver + "-rbxPkgManifest.txt")
//	pkgs := ConstructPackages(manif)
//
//	var wg sync.WaitGroup
//	wg.Add(len(pkgs))
//	
//	for _, pkg := range pkgs {
//		go func(ver string, pkg Package) {
//			DownloadPackage(ver, "cache", pkg)
//			ExtractPackage("cache", "bleh", pkg)
//			wg.Done()
//		}(ver, pkg)
//	}
//
//	wg.Wait()
//}
