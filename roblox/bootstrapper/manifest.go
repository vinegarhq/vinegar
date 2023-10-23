package bootstrapper

import (
	"fmt"
	"log"
	"strings"

	"github.com/vinegarhq/vinegar/roblox/version"
	"github.com/vinegarhq/vinegar/util"
)

type Manifest struct {
	*version.Version
	DeployURL string
	Packages
}

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
