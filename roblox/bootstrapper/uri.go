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

func ParseStudioURI(launchURI string) (args []string) {
	tempArgMap := make(map[string]string, 0)

	for _, param := range strings.Split(launchURI, "+") {
		pair := strings.Split(param, ":")
		if len(pair) != 2 {
			continue
		}

		key, val := pair[0], pair[1]

		if key == "gameinfo" {
			args = append(args,
				"-url", "https://www.roblox.com/Login/Negotiate.ashx", "-ticket", val,
			)
		} else {
			tempArgMap[key] = val
			args = append(args, "-"+key, "val")
		}
	}

	if tempArgMap["launchmode"] != "" && tempArgMap["task"] == "" {
		switch tempArgMap["launchmode"] {
		case "plugin":
			args = append(args, "-task", "InstallPlugin", "-pluginId", tempArgMap["pluginid"])
		case "asset":
			args = append(args, "-task", "TryAsset", "-assetId", tempArgMap["pluginid"])
		case "edit":
			args = append(args, "-task", "EditPlace")
		}
	}

	return
}
