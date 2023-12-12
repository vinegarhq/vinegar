package state

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/roblox/version"
)

func TestState(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "state")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Load()
	if err != nil && !reflect.DeepEqual(s, State{}) {
		t.Fatal("want empty state on no file")
	}

	path = f.Name()

	s, err = Load()
	if err != nil {
		t.Fatal(err)
	}

	v := version.New(roblox.Player, "", "version-meowmeowmrrp")
	s.DxvkVersion = "6.9"
	s.AddBinary(&bootstrapper.PackageManifest{
		Version: &v,
		Packages: bootstrapper.Packages{{
			Checksum: "meow",
		}},
	})

	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	sExp, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	log.Println(sExp)
	if sExp.Version(roblox.Player) != v.GUID {
		t.Fatal("want version stored state")
	}

	if sExp.DxvkVersion != "6.9" {
		t.Fatal("want dxvk version stored state")
	}

	if !reflect.DeepEqual(sExp.Packages(), []string{"meow"}) {
		t.Fatal("want meow packages")
	}
}
