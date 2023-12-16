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

// defaultRenderer is used as the default renderer when
// no explicit named renderer argument has been given.
const DefaultRenderer = "D3D11"

var renderers = []string{
	"OpenGL",
	"D3D11FL10",
	"D3D11",
	"Vulkan",
}

// FFlags is Roblox's Fast Flags implemented in map form.
type FFlags map[string]interface{}

// Apply creates and compiles the FFlags file and
// directory in the named versionDir.
func (f FFlags) Apply(versionDir string) error {
	dir := filepath.Join(versionDir, "ClientSettings")
	path := filepath.Join(dir, "ClientAppSettings.json")

	log.Println("Applying custom FFlags:", path)

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

// ValidRenderer determines if the named renderer is part of
// the available supported Roblox renderer backends, used in
// SetRenderer.
func ValidRenderer(renderer string) bool {
	for _, r := range renderers {
		if renderer == r {
			return true
		}
	}

	return false
}

// SetRenderer sets the named renderer to the FFlags, by disabling
// all other unused renderers.
func (f FFlags) SetRenderer(renderer string) error {
	if renderer == "" {
		renderer = DefaultRenderer
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
