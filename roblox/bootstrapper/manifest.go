package bootstrapper

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/vinegarhq/vinegar/roblox/version"
	"github.com/vinegarhq/vinegar/util"
	"golang.org/x/sync/errgroup"
)

type Manifest struct {
	*version.Version
	DeployURL string
	Packages
}

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

func channelPath(channel string) string {
	// Ensure that the channel is lowercased, since internally in
	// ClientSettings it will be lowercased, but not on the deploy mirror.
	channel = strings.ToLower(channel)

	// Roblox CDN only accepts no channel if its the default channel
	if channel == "" || channel == version.DefaultChannel {
		return "/"
	}

	return "/channel/" + channel + "/"
}

func FetchManifest(ver *version.Version) (Manifest, error) {
	cdn, err := CDN()
	if err != nil {
		return Manifest{}, err
	}
	durl := cdn + channelPath(ver.Channel) + ver.GUID

	log.Printf("Fetching manifest for %s (%s)", ver.GUID, durl)

	manif, err := util.Body(durl + "-rbxPkgManifest.txt")
	if err != nil {
		return Manifest{}, fmt.Errorf("fetch %s manifest: %w", ver.GUID, err)
	}

	pkgs, err := ParsePackages(strings.Split(manif, "\r\n"))
	if err != nil {
		return Manifest{}, err
	}

	return Manifest{
		Version:   ver,
		DeployURL: durl,
		Packages:  pkgs,
	}, nil
}

func ParsePackages(manifest []string) (Packages, error) {
	pkgs := make(Packages, 0)

	if (len(manifest)-1)%4 != 0 {
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
