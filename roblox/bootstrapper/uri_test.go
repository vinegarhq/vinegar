package bootstrapper

import (
	"testing"
)

func TestProtocolURIParse(t *testing.T) {
	mime := "roblox-player:1+launchmode:play"
	puri := ParseProtocolURI(mime)

	if puri["launchmode"] != "play" {
		t.Fatalf("want launchmode play, got %s", puri["launchmode"])
	}
}
