package version

import (
	"log"
	"strings"

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

func ChannelPath(channel string) string {
	// Ensure that the channel is lowercased, since internally in
	// ClientSettings it will be lowercased, but not on the deploy mirror.
	channel = strings.ToLower(channel)

	if channel == DefaultChannel {
		return "/"
	}

	return "/channel/" + channel + "/"
}

func New(bt roblox.BinaryType, channel string, GUID string) Version {
	log.Printf("Got %s version %s", bt.BinaryName(), GUID)

	return Version{
		Type:    bt,
		Channel: channel,
		GUID:    GUID,
	}
}

func Fetch(bt roblox.BinaryType, channel string) (Version, error) {
	c := ""
	if c != "" {
		c = "for channel " + channel
	}
	log.Printf("Fetching latest version of %s %s", bt.BinaryName(), c)

	cv, err := api.GetClientVersion(bt, channel)
	if err != nil {
		return Version{}, err
	}

	log.Printf("Fetched %s canonical version %s", bt.BinaryName(), cv.Version)

	return New(bt, channel, cv.ClientVersionUpload), nil
}
