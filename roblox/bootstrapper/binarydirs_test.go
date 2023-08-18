package bootstrapper

import (
	"reflect"
	"testing"
	"github.com/vinegarhq/aubun/roblox"
)

func TestProperBinaryDirectories(t *testing.T) {
	if !reflect.DeepEqual(Directories(roblox.Player), PlayerDirectories) {
		t.Fatal("Player's directories is not the player's stored package directories")
	}
	if !reflect.DeepEqual(Directories(roblox.Studio), StudioDirectories) {
		t.Fatal("Player's directories is not the player's stored package directories")
	}
	if Directories(roblox.BinaryType(-1)) != nil {
		t.FailNow()
	}
}
