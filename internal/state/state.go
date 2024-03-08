package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/vinegarhq/vinegar/internal/dirs"
)

var path = filepath.Join(dirs.Data, "state.json")

// BinaryState is used track a Binary's deployment and wineprefix.
type Binary struct {
	DxvkVersion string
	Version     string
	Packages    []string
}

// State holds various details about Vinegar's current state.
type State struct {
	Player Binary
	Studio Binary
}

// Load returns the state file's contents in State form.
//
// If the state file does not exist or is empty, an
// empty state is returned.
func Load() (State, error) {
	var state State

	f, err := os.ReadFile(path)
	if (err != nil && errors.Is(err, os.ErrNotExist)) || len(f) == 0 {
		return State{}, nil
	}
	if err != nil {
		return State{}, err
	}

	if err := json.Unmarshal(f, &state); err != nil {
		return State{}, err
	}

	return state, nil
}

// Save saves the current state to the state file.
func (s *State) Save() error {
	if err := dirs.Mkdirs(filepath.Dir(path)); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	state, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}

	if _, err := f.Write(state); err != nil {
		return err
	}

	return nil
}

// Packages returns all the available Binary packages from the state.
func (s *State) Packages() (pkgs []string) {
	for _, bs := range []Binary{s.Player, s.Studio} {
		pkgs = append(pkgs, bs.Packages...)
	}

	return
}

// Packages returns all the available Binary versions from the state.
func (s *State) Versions() (vers []string) {
	for _, bs := range []Binary{s.Player, s.Studio} {
		vers = append(vers, bs.Version)
	}

	return
}
