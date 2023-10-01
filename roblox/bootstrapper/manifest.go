package bootstrapper

import (
	"os"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/util"
)

const ManifestSuffix = "-rbxPkgManifest.txt"

type Manifest struct {
	roblox.Version
	DownloadDir string
	DeployURL   string
	Packages
}

func Fetch(ver roblox.Version, downloadDir string) (Manifest, error) {
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		return Manifest{}, err
	}

	cdn, err := CDN()
	if err != nil {
		return Manifest{}, err
	}

	deployURL := cdn + roblox.ChannelPath(ver.Channel) + ver.GUID

	log.Printf("Fetching latest manifest for %s (%s)", ver.GUID, deployURL)

	manifest, err := util.Body(deployURL + ManifestSuffix)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to fetch manifest: %w, is your channel valid?", err)
	}

	pkgs, err := ParsePackages(strings.Split(manifest, "\r\n"))
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return Manifest{
		Version:     ver,
		DownloadDir: downloadDir,
		DeployURL:   deployURL,
		Packages:    pkgs,
	}, nil
}

func (m *Manifest) Download() error {
	log.Printf("Downloading %d Packages", len(m.Packages))

	return m.Packages.Perform(func(pkg Package) error {
		return pkg.Fetch(filepath.Join(m.DownloadDir, pkg.Checksum), m.DeployURL)
	})
}

func (m *Manifest) Extract(dir string, dirs PackageDirectories) error {
	log.Printf("Extracting %d Packages", len(m.Packages))

	return m.Packages.Perform(func(pkg Package) error {
		dest, ok := dirs[pkg.Name]

		if !ok {
			return fmt.Errorf("unhandled package: %s", pkg.Name)
		}

		return pkg.Extract(
			filepath.Join(m.DownloadDir, pkg.Checksum),
			filepath.Join(dir, dest),
		)
	})
}

func (m *Manifest) Setup(dir string, dirs PackageDirectories) error {
	if err := m.Download(); err != nil {
		return err
	}

	if err := m.Extract(dir, dirs); err != nil {
		return err
	}

	return WriteAppSettings(dir)
}
