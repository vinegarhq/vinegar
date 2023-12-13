// Package version implements version handling for a Roblox binary.
package version

import (
	"log"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/api"
)

// DefaultChannel is the default channel used for when
// no channel has been given to New or Fetch.
const DefaultChannel = "live"

// Version is a representation of a Binary's version
type Version struct {
	Type    roblox.BinaryType
	Channel string
	GUID    string
}

// New returns a new Version.
func New(bt roblox.BinaryType, channel string, GUID string) Version {
	if channel == "" {
		channel = DefaultChannel
	}

	log.Printf("Got %s version %s with channel %s", bt.BinaryName(), GUID, channel)

	return Version{
		Type:    bt,
		Channel: channel,
		GUID:    GUID,
	}
}

// Fetch retrieves the latest Version for the [roblox.BinaryType] with the deployment channel
// through [api.GetClientVersion].
func Fetch(bt roblox.BinaryType, channel string) (Version, error) {
	if channel == "" {
		channel = DefaultChannel
	}

	log.Printf("Fetching latest version of %s for channel %s", bt.BinaryName(), channel)

	cv, err := api.GetClientVersion(bt.BinaryName(), channel)
	if err != nil {
		return Version{}, err
	}

	log.Printf("Fetched %s canonical version %s", bt.BinaryName(), cv.Version)

	return New(bt, channel, cv.ClientVersionUpload), nil
}
