package bootstrapper

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/util"
)

const ManifestSuffix = "-rbxPkgManifest.txt"

type Manifest struct {
	roblox.Version
	SourceDir string
	Packages
}

func FetchManifest(ver roblox.Version, srcDir string) (Manifest, error) {
	log.Printf("Fetching latest manifest for %s (%s)", ver.GUID, ver.DeployURL)

	manifest, err := util.Body(ver.DeployURL + ManifestSuffix)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to fetch manifest: %w, is your channel valid?", err)
	}

	pkgs, err := ParsePackages(strings.Split(manifest, "\r\n"))
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return Manifest{
		Version:   ver,
		SourceDir: srcDir,
		Packages:  pkgs,
	}, nil
}

func (m *Manifest) Download() error {
	log.Printf("Downloading %d Packages", len(m.Packages))

	return m.Packages.Perform(func(pkg Package) error {
		url := m.Version.DeployURL + "-" + pkg.Name
		dest := filepath.Join(m.SourceDir, pkg.Checksum)

		if _, err := os.Stat(dest); err == nil {
			log.Printf("Package %s is already downloaded", pkg.Name)
			return util.VerifyFileMD5(dest, pkg.Checksum)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		if err := util.Download(url, dest); err != nil {
			log.Printf("Unable to download package %s, retrying...", pkg.Name)

			if err := util.Download(url, dest); err != nil {
				return fmt.Errorf("failed to download package %s: %w", pkg.Name, err)
			}
		}

		log.Printf("Downloaded Package %s", pkg.Name)
		return util.VerifyFileMD5(dest, pkg.Checksum)
	})
}

func (m *Manifest) Extract(dir string, dirs PackageDirectories) error {
	log.Printf("Extracting %d Packages", len(m.Packages))

	err := m.Packages.Perform(func(pkg Package) error {
		src := filepath.Join(m.SourceDir, pkg.Checksum)
		dest, ok := dirs[pkg.Name]

		if !ok {
			return fmt.Errorf("unhandled package: %s", pkg.Name)
		}

		if err := util.Extract(src, filepath.Join(dir, dest)); err != nil {
			return fmt.Errorf("failed to extract package %s: %w", pkg.Name, err)
		}

		log.Printf("Extracted Package %s (%s)", pkg.Name, pkg.Checksum)
		return nil
	})
	if err != nil {
		return err
	}

	return WriteAppSettings(dir)
}
