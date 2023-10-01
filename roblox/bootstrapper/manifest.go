package bootstrapper

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/util"
)

const ManifestSuffix = "-rbxPkgManifest.txt"

type Manifest struct {
	roblox.Version
	DeployURL string
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
		Version:   ver,
		DeployURL: deployURL,
		Packages:  pkgs,
	}, nil
}
