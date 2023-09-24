package bootstrapper

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/vinegarhq/vinegar/util"
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

func PackageExcluded(name string) bool {
	for _, ex := range ExcludedPackages {
		if name == ex {
			return true
		}
	}

	return false
}

func (p *Package) Verify(src string) error {
	log.Printf("Verifying Package %s (%s)", p.Name, p.Checksum)

	pkgFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer pkgFile.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, pkgFile); err != nil {
		return err
	}

	if p.Checksum != hex.EncodeToString(hash.Sum(nil)) {
		return fmt.Errorf("package %s is corrupted")
	}

	return nil
}

func (p *Package) Download(dest, deployURL string) error {
	if err := p.Verify(dest); err == nil {
		log.Printf("Package %s is already downloaded", p.Name)
		return nil
	}

	log.Printf("Downloading Package %s", p.Name)
	if err := util.Download(deployURL+"-"+p.Name, dest); err != nil {
		return fmt.Errorf("failed to download package %s: %w", p.Name, err)
	}

	return p.Verify(dest)
}

func (p *Package) Fetch(dest, deployURL string) error {
	err := p.Download(dest, deployURL)
	if err == nil {
		return nil
	}

	log.Printf("Failed to fetch package %s: %s, retrying...", p.Name, err)

	return p.Download(dest, deployURL)
}

func (p *Package) Extract(src, dest string) error {
	if err := util.Extract(src, dest); err != nil {
		return fmt.Errorf("failed to extract package %s (%s): %w", p.Name, src, err)
	}

	log.Printf("Extracted Package %s (%s)", p.Name, p.Checksum)
	return nil
}
