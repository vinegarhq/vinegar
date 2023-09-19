package state

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/wine"
)

var pathFilename = "state.toml"

type ApplicationState struct {
	Version  string
	Packages []string
}

type ApplicationStates map[string]ApplicationState

type State struct {
	DxvkVersion  string
	Applications ApplicationStates
}

func GetPaths(pfx *wine.Prefix) (string, string) {
	var prefixDataPath = dirs.GetPrefixData(pfx)
	var statePath = filepath.Join(prefixDataPath, pathFilename)

	return prefixDataPath, statePath
}

func Save(pfx *wine.Prefix, state *State) error {
	var prefixDataPath, path = GetPaths(pfx)

	err := dirs.Mkdirs(prefixDataPath)
	if err != nil {
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

	err = toml.NewEncoder(file).Encode(state)
	if err != nil {
		return err
	}

	return nil
}

func Load(pfx *wine.Prefix) (State, error) {
	var _, path = GetPaths(pfx)
	var state State

	_, err := toml.DecodeFile(path, &state)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return State{}, err
	}

	return state, nil
}

func SaveManifest(pfx *wine.Prefix, manif *bootstrapper.Manifest) error {
	name := manif.Version.Type.String()

	log.Printf("Saving Manifest State for %s", name)

	state, err := Load(pfx)
	if err != nil {
		return err
	}

	app := ApplicationState{
		Version: manif.Version.GUID,
	}
	for _, pkg := range manif.Packages {
		app.Packages = append(app.Packages, pkg.Checksum)
	}

	if state.Applications == nil {
		state.Applications = make(ApplicationStates, 0)
	}

	state.Applications[name] = app

	return Save(pfx, &state)
}

func SaveDxvk(pfx *wine.Prefix, ver string) error {
	log.Printf("Saving installed DXVK State")

	state, err := Load(pfx)
	if err != nil {
		return err
	}

	state.DxvkVersion = ver

	return Save(pfx, &state)
}

func Packages(pfx *wine.Prefix) ([]string, error) {
	var packages []string

	states, err := Load(pfx)
	if err != nil {
		return []string{}, err
	}

	for _, info := range states.Applications {
		packages = append(packages, info.Packages...)
	}

	return packages, nil
}

func Version(pfx *wine.Prefix, bt roblox.BinaryType) (string, error) {
	states, err := Load(pfx)
	if err != nil {
		return "", err
	}

	return states.Applications[bt.String()].Version, nil
}

func Versions(pfx *wine.Prefix) ([]string, error) {
	var versions []string

	states, err := Load(pfx)
	if err != nil {
		return []string{}, err
	}

	for _, info := range states.Applications {
		versions = append(versions, info.Version)
	}

	return versions, nil
}

func ClearApplications(pfx *wine.Prefix) error {
	state, err := Load(pfx)
	if err != nil {
		return err
	}

	state.Applications = nil

	return Save(pfx, &state)
}

func DxvkVersion(pfx *wine.Prefix) (string, error) {
	states, err := Load(pfx)
	if err != nil {
		return "", err
	}

	return states.DxvkVersion, nil
}
