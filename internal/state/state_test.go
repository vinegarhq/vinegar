package state

import (
	"os"
	"reflect"
	"testing"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/bootstrapper"
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

	v := bootstrapper.NewDeployment(roblox.Player, "", "version-meowmeowmrrp")
	s.Player.Add(&bootstrapper.PackageManifest{
		Deployment: &v,
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

	if sExp.Player.Version != v.GUID {
		t.Fatal("want version stored state")
	}

	if !reflect.DeepEqual(sExp.Packages(), []string{"meow"}) {
		t.Fatal("want meow packages")
	}
}
