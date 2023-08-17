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
	ExcludedPackages            = []string{
		"RobloxPlayerLauncher.exe",
		"WebView2RuntimeInstaller.zip",
	}
)

type Package struct {
	Name     string
	Checksum string
	Size     int64
}

type Packages []Package

func ParsePackages(manif []string) (Packages, error) {
	pkgs := make(Packages, 0)

	if len(manif) < 5 {
		return pkgs, ErrInvalidManifest
	}

	if manif[0] != "v0" {
		return pkgs, fmt.Errorf("%w: %s", ErrUnhandledManifestVersion, manif[0])
	}

	for i := 1; i <= len(manif)-4; i += 4 {
		if PackageExcluded(manif[i]) {
			continue
		}

		size, err := strconv.ParseInt(manif[i+3], 10, 64)
		if err != nil {
			return pkgs, err
		}

		pkgs = append(pkgs, Package{
			Name:     manif[i],
			Checksum: manif[i+1],
			Size:     size,
		})
	}

	return pkgs, nil
}

func (pkgs Packages) Perform(fn func(Package) error) error {
	var eg errgroup.Group

	for _, pkg := range pkgs {
		pkg := pkg

		eg.Go(func() error {
			return fn(pkg)
		})
	}

	return eg.Wait()
}

func PackageExcluded(name string) bool {
	for _, ex := range ExcludedPackages {
		if name == ex {
			return true
		}
	}

	return false
}
