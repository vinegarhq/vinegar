package version

import (
	"log"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/api"
)

const DefaultChannel = "live"

type ClientVersion struct {
	Version                 string `json:"version"`
	ClientVersionUpload     string `json:"clientVersionUpload"`
	BootstrapperVersion     string `json:"bootstrapperVersion"`
	NextClientVersionUpload string `json:"nextClientVersionUpload"`
	NextClientVersion       string `json:"nextClientVersion"`
}

type Version struct {
	Type    roblox.BinaryType
	Channel string
	GUID    string
}

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

func Fetch(bt roblox.BinaryType, channel string) (Version, error) {
	if channel == "" {
		channel = DefaultChannel
	}

	log.Printf("Fetching latest version of %s for channel %s", bt.BinaryName(), channel)

	cv, err := api.GetClientVersion(bt, channel)
	if err != nil {
		return Version{}, err
	}

	log.Printf("Fetched %s canonical version %s", bt.BinaryName(), cv.Version)

	return New(bt, channel, cv.ClientVersionUpload), nil
}
