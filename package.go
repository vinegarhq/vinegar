package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinegarhq/vinegar/util"
	"golang.org/x/sync/errgroup"
)

type Package struct {
	Name     string
	URL      string
	Checksum string
}

type Packages []Package

func GetPackages(url string) Packages {
	log.Printf("Fetching packages for %s", url)

	rawManif, err := util.URLBody(url + "-rbxPkgManifest.txt")
	if err != nil {
		log.Fatal(err)
	}

	manif := strings.Split(string(rawManif), "\r\n")

	if manif[0] != "v0" {
		log.Fatal(err)
	}

	var pkgs Packages
	for i := 1; i < len(manif)-4; i += 4 {
		if IsExcluded(manif[i]) {
			continue
		}

		pkgs = append(pkgs, Package{
			manif[i],
			url + "-" + manif[i],
			manif[i+1],
		})
	}

	return pkgs
}

func IsExcluded(name string) bool {
	for _, ex := range []string{
		"RobloxPlayerLauncher.exe",
		"WebView2RuntimeInstaller.zip",
	} {
		if name == ex {
			return true
		}
	}

	return false
}

// The usage of errgroup is always preferred over waitgroups, since Vinegar's
// error handling is not return handled, and will always stop Vinegar's execution.

func (ps Packages) Download() {
	eg := new(errgroup.Group)

	if err := os.MkdirAll(Dirs.Downloads, DirMode); err != nil {
		log.Fatal(err)
	}

	log.Println("Downloading all packages")
	for _, pkg := range ps {
		dst := filepath.Join(Dirs.Downloads, pkg.Checksum)
		pkg := pkg

		eg.Go(func() error {
			if _, err := os.Stat(dst); err != nil {
				if err := util.Download(pkg.URL, dst); err != nil {
					return err
				}
			}

			log.Printf("Verifying %s package %s", pkg.Checksum, pkg.Name)

			return util.Verify(dst, pkg.Checksum)
		})
	}

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}

func (ps Packages) Extract(dstDir string, dsts map[string]string) {
	eg := new(errgroup.Group)

	log.Println("Extracting all packages")
	if err := os.MkdirAll(dstDir, DirMode); err != nil {
		log.Fatal(err)
	}

	for _, pkg := range ps {
		src := filepath.Join(Dirs.Downloads, pkg.Checksum)

		if _, ok := dsts[pkg.Name]; !ok {
			log.Printf("Warning: unhandled package: %s", pkg.Name)
			continue
		}

		dst := filepath.Join(dstDir, dsts[pkg.Name])
		pkg := pkg

		eg.Go(func() error {
			err := util.Unzip(src, dst)
			if err != nil {
				return err
			}

			log.Printf("Extracted package %s", pkg.Name)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		log.Fatal("extract: ", err)
	}
}

func (ps Packages) Cleanup() {
	log.Println("Removing unused packages")

	pkgFiles, err := os.ReadDir(Dirs.Downloads)
	if err != nil {
		log.Fatal(err)
	}

find:
	for _, file := range pkgFiles {
		for _, pkg := range ps {
			if file.Name() == pkg.Checksum {
				continue find
			}
		}

		log.Printf("Removing %s", file.Name())
		if err := os.Remove(filepath.Join(Dirs.Downloads, file.Name())); err != nil {
			log.Fatal(err)
		}
	}
}
