package bootstrapper

import (
	"reflect"
	"testing"
)

func TestProperBinaryDirectories(t *testing.T) {
	if !reflect.DeepEqual(Player.Directories(), PlayerDirectories) {
		t.Fatal("Player's directories is not the player's stored package directories")
	}
	if !reflect.DeepEqual(Studio.Directories(), StudioDirectories) {
		t.Fatal("Player's directories is not the player's stored package directories")
	}
	if BinaryType(-1).Directories() != nil {
		t.FailNow()
	}
}
