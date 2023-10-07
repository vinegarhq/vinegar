package bootstrapper

import (
	"fmt"
	"log"
	"strings"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/util"
)

type Manifest struct {
	*roblox.Version
	DeployURL string
	Packages
}

func FetchManifest(ver *roblox.Version) (Manifest, error) {
	cdn, err := CDN()
	if err != nil {
		return Manifest{}, err
	}

	deployURL := cdn + roblox.ChannelPath(ver.Channel) + ver.GUID

	log.Printf("Fetching manifest for %s (%s)", ver.GUID, deployURL)

	manif, err := util.Body(deployURL + "-rbxPkgManifest.txt")
	if err != nil {
		return Manifest{}, fmt.Errorf("fetch %s manifest: %w, is your channel valid?", ver.GUID, err)
	}

	pkgs, err := ParsePackages(strings.Split(manif, "\r\n"))
	if err != nil {
		return Manifest{}, err
	}

	return Manifest{
		Version:   ver,
		DeployURL: deployURL,
		Packages:  pkgs,
	}, nil
}
