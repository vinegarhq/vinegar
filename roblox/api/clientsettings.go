package api

import (
	"github.com/vinegarhq/vinegar/roblox"
)

type ClientVersion struct {
	Version                 string `json:"version"`
	ClientVersionUpload     string `json:"clientVersionUpload"`
	BootstrapperVersion     string `json:"bootstrapperVersion"`
	NextClientVersionUpload string `json:"nextClientVersionUpload,omitempty"`
	NextClientVersion       string `json:"nextClientVersion,omitempty"`
}

func GetClientVersion(bt roblox.BinaryType, channel string) (ClientVersion, error) {
	var cv ClientVersion

	ep := "v2/client-version/" + bt.BinaryName()
	if channel != "" {
		ep += "/channel/" + channel
	}

	err := Request("GET", "clientsettings", ep, &cv)
	if err != nil {
		return ClientVersion{}, err
	}

	return cv, nil
}
