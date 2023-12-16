package bootstrapper

import (
	"log"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/api"
)

// DefaultChannel is the default channel used for when
// no named channel argument has been given.
const DefaultChannel = "LIVE"

// Version is a representation of a Binary's deployment or version.
type Deployment struct {
	Type    roblox.BinaryType
	Channel string
	GUID    string
}

// NewDeployment returns a new Deployment.
func NewDeployment(bt roblox.BinaryType, channel string, GUID string) Deployment {
	if channel == "" {
		channel = DefaultChannel
	}

	return Deployment{
		Type:    bt,
		Channel: channel,
		GUID:    GUID,
	}
}

// FetchDeployment returns the latest Version for the given roblox Binary type
// with the given deployment channel through [api.GetClientVersion].
func FetchDeployment(bt roblox.BinaryType, channel string) (Deployment, error) {
	if channel == "" {
		channel = DefaultChannel
	}

	log.Printf("Fetching latest version of %s for channel %s", bt.BinaryName(), channel)

	cv, err := api.GetClientVersion(bt.BinaryName(), channel)
	if err != nil {
		return Deployment{}, err
	}

	log.Printf("Fetched %s canonical version %s", bt.BinaryName(), cv.Version)

	return NewDeployment(bt, channel, cv.ClientVersionUpload), nil
}
