package bootstrapper

import (
	"log/slog"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/api"
)

// Version is a representation of a Binary's deployment or version.
//
// Channel can either be a given channel, or empty - in which Roblox
// will consider the 'default' channel.
//
// In all things related to the Roblox API, the default channel is empty,
// or 'live'/'LIVE' on clientsettings. On the Client/Studio, the default channel
// is (or can be) 'production'. This behavior is undocumented, so it's best to
// just use an empty channel i guess.
type Deployment struct {
	Type    roblox.BinaryType
	Channel string
	GUID    string
}

// NewDeployment returns a new Deployment.
func NewDeployment(bt roblox.BinaryType, channel string, GUID string) Deployment {
	return Deployment{
		Type:    bt,
		Channel: channel,
		GUID:    GUID,
	}
}

// FetchDeployment returns the latest Version for the given roblox Binary type
// with the given deployment channel through [api.GetClientVersion].
func FetchDeployment(bt roblox.BinaryType, channel string) (Deployment, error) {
	slog.Info("Fetching Binary Deployment", "name", bt.BinaryName(), "channel", channel)

	cv, err := api.GetClientVersion(bt.BinaryName(), channel)
	if err != nil {
		return Deployment{}, err
	}

	return NewDeployment(bt, channel, cv.ClientVersionUpload), nil
}
