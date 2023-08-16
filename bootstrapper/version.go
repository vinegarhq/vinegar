package bootstrapper

import (
	"strings"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/vinegarhq/aubun/util"
)

const VersionCheckURL = "https://clientsettingscdn.roblox.com/v2/client-version"

var (
	ErrNoCDNFound = errors.New("failed to find an accessible roblox deploy mirror")
	ErrNoVersion  = errors.New("no version found")
	CDNURLs       = []string{
		"https://setup.rbxcdn.com",
		"https://setup-ak.rbxcdn.com",
		"https://setup-cfly.rbxcdn.com",
		"https://s3.amazonaws.com/setup.roblox.com",
	}
)

type ClientVersion struct {
	Version                 string `json:"version"`
	ClientVersionUpload     string `json:"clientVersionUpload"`
	BootstrapperVersion     string `json:"bootstrapperVersion"`
	NextClientVersionUpload	string `json:"nextClientVersionUpload"`
	NextClientVersion	    string `json:"nextClientVersion"`
}

type Version struct {
	Type BinaryType
	URL  string
	GUID string
}

func FindCDN() (string, error) {
	log.Println("Finding an accessible roblox deploy mirror")

	for _, _cdn := range CDNURLs {
		resp, err := http.Head(_cdn + "/" + "version")

		if err == nil && resp.StatusCode == 200 {
			log.Printf("Found mirror: %s", _cdn)

			return _cdn, nil
		}
	}

	return "", ErrNoCDNFound
}

func LatestVersion(bt BinaryType, channel string) (Version, error) {
	if channel == "" {
		channel = "LIVE"
	}

	// deploy mirror doesnt like uppercase channel names
	suf := "/"
	if channel != "" && channel != "LIVE" {
		suf += "channel/" + strings.ToLower(channel) + "/"
	}

	log.Printf("Fetching latest version of %s for channel %s", bt.String(), channel)

	resp, err := util.Body(VersionCheckURL + "/" + bt.String() + "/channel/" + channel)
	if err != nil {
		return Version{}, fmt.Errorf("failed to fetch version: %w", err)
	}

	var cv ClientVersion
	if err := json.Unmarshal([]byte(resp), &cv); err != nil {
		return Version{}, fmt.Errorf("failed to parse clientsettings: %w", err)
	}

	if cv.ClientVersionUpload == "" {
		return Version{}, ErrNoVersion
	}

	log.Printf("Fetched %s version: %s (%s)", bt.String(), cv.Version, cv.ClientVersionUpload)

	cdn, err := FindCDN()
	if err != nil {
		return Version{}, err
	}

	return Version{
		Type: bt,
		URL:  cdn + suf + cv.ClientVersionUpload,
		GUID: cv.ClientVersionUpload,
	}, nil
}
