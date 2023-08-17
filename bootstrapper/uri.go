package bootstrapper

import (
	"strings"
	"net/url"
)

var PlayerURIKeyFlags = map[string]string{
	"launchmode":       "--",
	"gameinfo":         "-t ",
	"placelauncherurl": "-j ",
	"launchtime":       "--launchtime=",
	"browsertrackerid": "-b ",
	"robloxLocale":     "--rloc ",
	"gameLocale":       "--gloc ",
	"channel":          "-channel ",
}

func ParsePlayerURI(launchURI string) (string, []string) {
	channel := ""
	args := make([]string, 0)

	for _, uri := range strings.Split(launchURI, "+") {
		parts := strings.Split(uri, ":")

		if len(parts) != 2 || URIMap[parts[0]] == "" {
			continue
		}

		if parts[0] == "launchmode" && parts[1] == "play" {
			parts[1] = "app"
		}

		if parts[0] == "channel" {
			channel = strings.ToLower(parts[1])
		}

		if parts[0] == "placelauncherurl" {
			urlDecoded, _ := url.QueryUnescape(parts[1])
			parts[1] = urlDecoded
		}

		args = append(args, URIMap[parts[0]]+parts[1])
	}

	return channel, args
}
