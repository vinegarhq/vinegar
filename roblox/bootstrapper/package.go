package bootstrapper

import (
	"errors"
	"fmt"
	"strconv"

	"golang.org/x/sync/errgroup"
)

var (
	ErrInvalidManifest          = errors.New("invalid package manifest given")
	ErrUnhandledManifestVersion = errors.New("unhandled package manifest version")
)

type Package struct {
	Name     string
	Checksum string
	Size     int64
}

type Packages []Package

func ParsePackages(manifest []string) (Packages, error) {
	pkgs := make(Packages, 0)

	if len(manifest) < 5 {
		return pkgs, ErrInvalidManifest
	}

	if manifest[0] != "v0" {
		return pkgs, fmt.Errorf("%w: %s", ErrUnhandledManifestVersion, manifest[0])
	}

	for i := 1; i <= len(manifest)-4; i += 4 {
		if manifest[i] == "RobloxPlayerLauncher.exe" ||
			manifest[i] == "WebView2RuntimeInstaller.zip" {
			continue
		}

		size, err := strconv.ParseInt(manifest[i+3], 10, 64)
		if err != nil {
			return pkgs, err
		}

		pkgs = append(pkgs, Package{
			Name:     manifest[i],
			Checksum: manifest[i+1],
			Size:     size,
		})
	}

	return pkgs, nil
}

func (pkgs *Packages) Perform(fn func(Package) error) error {
	var eg errgroup.Group

	for _, pkg := range *pkgs {
		pkg := pkg

		eg.Go(func() error {
			return fn(pkg)
		})
	}

	return eg.Wait()
}
