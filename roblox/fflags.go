package roblox

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	DefaultRenderer = "D3D11"
	Renderers       = []string{
		"OpenGL",
		"D3D11FL10",
		DefaultRenderer,
		"Vulkan",
	}
)

type FFlags map[string]interface{}

func (f *FFlags) Apply(versionDir string) error {
	dir := filepath.Join(versionDir, "ClientSettings")
	path := filepath.Join(dir, "ClientAppSettings.json")

	// If the fflags are empty, the FFlags file's contents will be 'null'
	if len(*f) == 0 {
		return nil
	}

	log.Printf("Applying custom FFlags")

	err := os.Mkdir(dir, 0o755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer file.Close()

	fflags, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("FFlags used: %s", string(fflags))

	_, err = file.Write(fflags)
	if err != nil {
		return err
	}

	return nil
}

func ValidRenderer(renderer string) bool {
	if renderer == "" {
		renderer = DefaultRenderer
	}

	for _, r := range Renderers {
		if renderer == r {
			return true
		}
	}

	return false
}

func (f *FFlags) SetRenderer(renderer string) error {
	if renderer == "" {
		renderer = DefaultRenderer
	}

	if !ValidRenderer(renderer) {
		return fmt.Errorf("invalid renderer given: %s", renderer)
	}

	if len(*f) == 0 {
		*f = make(FFlags)
	}

	log.Printf("Using renderer: %s", renderer)

	// Disable all other renderers except the given one.
	for _, r := range Renderers {
		isRenderer := r == renderer

		(*f)["FFlagDebugGraphicsPrefer"+r] = isRenderer
		(*f)["FFlagDebugGraphicsDisable"+r] = !isRenderer
	}

	return nil
}
