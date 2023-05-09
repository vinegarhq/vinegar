package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

//go:embed rco.json
var rawRCO []byte

func (r *Roblox) ApplyFFlags(app string) {
	fflagsDir := filepath.Join(r.VersionDir, app+"Settings")

	if err := os.MkdirAll(fflagsDir, DirMode); err != nil {
		log.Fatal(err)
	}

	fflagsFile, err := os.Create(filepath.Join(fflagsDir, app+"AppSettings.json"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Applying custom FFlags to %s", app)

	Config.SetFFlags()

	fflagsJSON, err := json.MarshalIndent(Config.FFlags, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := fflagsFile.Write(fflagsJSON); err != nil {
		log.Fatal(err)
	}
}

func (c *Configuration) SetFFlags() {
	if c.RCO {
		c.SetRCOFFlags()
	}

	if c.Renderer != "" {
		c.SetFFlagRenderer()
	}
}

// 'Appends' RCO's FFlags to the global configuration, without
// overwriting the user-set FFlags.
func (c *Configuration) SetRCOFFlags() {
	log.Println("Applying RCO FFlags")

	var rco map[string]interface{}

	if err := json.Unmarshal(rawRCO, &rco); err != nil {
		log.Fatal(err)
	}

	for key, val := range c.FFlags {
		rco[key] = val
	}

	c.FFlags = rco
}

// Disable all given renderers and only enable the one given in the
// configuration, also validate it if it doesnt match the set of available
// renderers.
func (c *Configuration) SetFFlagRenderer() {
	possibleRenderers := []string{
		"OpenGL",
		"D3D11FL10",
		"D3D11",
		"Vulkan",
	}

	validRenderer := false

	for _, r := range possibleRenderers {
		if c.Renderer == r {
			validRenderer = true
		}
	}

	if !validRenderer {
		log.Fatal("invalid renderer, must be one of:", possibleRenderers)
	}

	for _, r := range possibleRenderers {
		isRenderer := r == c.Renderer
		c.FFlags["FFlagDebugGraphicsPrefer"+r] = isRenderer
		c.FFlags["FFlagDebugGraphicsDisable"+r] = !isRenderer
	}
}
