package state

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/bootstrapper"
)

var path = filepath.Join(dirs.PrefixData, "state.toml")

// ApplicationState is used track a Binary's version and it's packages
type BinaryState struct {
	Version  string
	Packages []string
}

// ApplicationStates is a map representation with the string
// type being the binary name in string form.
type BinaryStates map[string]BinaryState

// State holds various details about Vinegar's configuration
type State struct {
	DxvkVersion  string
	Applications BinaryStates // called Applications to retain compatibility
}

// Load will load the state file in dirs.PrefixData and return it's
// contents. If the state file does not exist, it will return an
// empty state.
func Load() (State, error) {
	var state State

	_, err := toml.DecodeFile(path, &state)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return State{}, err
	}

	if state.Applications == nil {
		state.Applications = make(BinaryStates, 0)
	}

	return state, nil
}

// Save will save the state to a toml-encoded file in dirs.PrefixData
func (s *State) Save() error {
	if err := dirs.Mkdirs(filepath.Dir(path)); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("# State saved by vinegar. DO NOT EDIT DIRECTLY.\n\n")
	if err != nil {
		return err
	}

	return toml.NewEncoder(file).Encode(s)
}

// AddBinary adds a given package manifest's packages and it's checksums
// to the state's Applications, with the identifier as the package
// manifest's binary name.
func (s *State) AddBinary(pm *bootstrapper.PackageManifest) {
	b := BinaryState{
		Version: pm.Deployment.GUID,
	}
	for _, pkg := range pm.Packages {
		b.Packages = append(b.Packages, pkg.Checksum)
	}

	s.Applications[pm.Deployment.Type.BinaryName()] = b
}

// Packages retrieves all the available Binary packages from the state
func (s *State) Packages() (pkgs []string) {
	for _, info := range s.Applications {
		pkgs = append(pkgs, info.Packages...)
	}

	return
}

// Packages retrieves all the available Binary versions from the state
func (s *State) Versions() (vers []string) {
	for _, ver := range s.Applications {
		vers = append(vers, ver.Version)
	}

	return
}

// Version is retrieves the version of a Binary from the state
func (s *State) Version(bt roblox.BinaryType) string {
	return s.Applications[bt.BinaryName()].Version
}
