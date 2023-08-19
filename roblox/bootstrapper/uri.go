package bootstrapper

import (
	"net/url"
	"strings"
)

var PlayerURIKeyFlags = map[string]string{
	"gameinfo":         "--authenticationTicket",
	"placelauncherurl": "--joinScriptUrl",
	"launchtime":       "--launchtime",
	"browsertrackerid": "--browserTrackerId",
	"robloxLocale":     "--rloc",
	"gameLocale":       "--gloc",
}

func ParsePlayerURI(launchURI string) (args []string, channel string) {
	// Roblox Client forces usage of the desktop app
	args = append(args, "--app")

	for _, param := range strings.Split(launchURI, "+") {
		pair := strings.Split(param, ":")
		if len(pair) != 2 {
			continue
		}

		key, val := pair[0], pair[1]

		if key == "channel" {
			channel = val

			continue
		}

		flag, ok := PlayerURIKeyFlags[key]
		if !ok || val == "" {
			continue
		}

		if key == "placelauncherurl" {
			urlDecoded, _ := url.QueryUnescape(val)
			val = urlDecoded
		}

		// arguments are given as --flag value
		args = append(args, flag, val)
	}

	return
}
