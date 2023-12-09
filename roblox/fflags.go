package roblox

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var ErrInvalidRenderer = errors.New("invalid renderer given")

var renderers = []string{
	"OpenGL",
	"D3D11FL10",
	"D3D11",
	"Vulkan",
}

type FFlags map[string]interface{}

func (f FFlags) Apply(versionDir string) error {
	dir := filepath.Join(versionDir, "ClientSettings")
	path := filepath.Join(dir, "ClientAppSettings.json")

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

	_, err = file.Write(fflags)
	if err != nil {
		return err
	}

	return nil
}

func ValidRenderer(renderer string) bool {
	// Assume Roblox's internal default renderer
	if renderer == "" {
		return true
	}

	for _, r := range renderers {
		if renderer == r {
			return true
		}
	}

	return false
}

func (f FFlags) SetRenderer(renderer string) error {
	// Assume Roblox's internal default renderer
	if renderer == "" {
		return nil
	}

	if !ValidRenderer(renderer) {
		return fmt.Errorf("fflags: %w: %s", ErrInvalidRenderer, renderer)
	}

	// Disable all other renderers except the given one.
	for _, r := range renderers {
		isRenderer := r == renderer

		f["FFlagDebugGraphicsPrefer"+r] = isRenderer
		f["FFlagDebugGraphicsDisable"+r] = !isRenderer
	}

	return nil
}
