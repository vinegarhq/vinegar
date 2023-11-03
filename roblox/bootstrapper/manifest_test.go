package bootstrapper

import (
	"reflect"
	"testing"
)

func TestParseFiles(t *testing.T) {
	manifest := []string{
		"meow\\mew\\nyao",
		"026b271a21b03f2e564c036525356db5",
		"purr",
		"4d9ec7b52a29c80f3ce1f6a65b14b563",
		"MicrosoftEdgeWebview2Setup.exe",
		"610b1b60dc8729bad759c92f82ee2804",
	}

	fs, err := ParseFiles(manifest)
	if err != nil {
		t.Fatal(err)
	}

	fsWant := Manifest{
		File{
			Path:     "meow/mew/nyao",
			Checksum: "026b271a21b03f2e564c036525356db5",
		},
		File{
			Path:     "purr",
			Checksum: "4d9ec7b52a29c80f3ce1f6a65b14b563",
		},
	}

	if !reflect.DeepEqual(fs, fsWant) {
		t.Fail()
	}
}
