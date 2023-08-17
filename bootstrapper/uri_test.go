package bootstrapper

import (
	"testing"
)

func TestPlayerURIParsed(t *testing.T) {
	uri := "roblox-player:1"
	uri += "+launchmode:play"
	uri += "+gameinfo:token"
	uri += "+launchtime:0"
	uri += "+placelauncherurl:https%3A%2F%2Fassetgame.roblox.com%2Fgame%2FPlaceLauncher.ashx"
	uri += "+browsertrackerid:0"
	uri += "+robloxLocale:en_us"
	uri += "+gameLocale:en_us"
	uri += "+channel:Ganesh"

	argsWant := []string{
		"--app",
		"--authenticationTicket", "token",
		"--launchtime", "0",
		"--joinScriptUrl", "https://assetgame.roblox.com/game/PlaceLauncher.ashx",
		"--browserTrackerId", "0",
		"--rloc", "en_us",
		"--gloc", "en_us",
		"-channel", "Ganesh",
	}
	channelWant := "Ganesh"

	args, channel := ParsePlayerURI(uri)

	for i, val := range args {
		if val != argsWant[i] {
			t.Fatalf("launch player uri parsing failed, key %s, want key match for %s", val, argsWant[i])
		}
	}

	if channel != channelWant {
		t.Fatalf("launch player uri parsing failed, %v, want channel match for %s", channel, channelWant)
	}
}
