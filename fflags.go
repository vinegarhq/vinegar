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

// Validate the given renderer, and apply it to the given map (fflags);
// It will also disable every other renderer.
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

// Apply the configuration's FFlags to Roblox's FFlags file, named after app:
// ClientAppSettings.json, we also set (and check) the renderer specified in
// the configuration, then indent it to look pretty and write.
func (r *Roblox) ApplyFFlags(app string) {
	fflagsDir := filepath.Join(r.VersionDir, app+"Settings")

	if err := os.MkdirAll(fflagsDir, 0o755); err != nil {
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

// 'Append' RCO's FFlags to the given map pointer.
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

func (c *Configuration) SetFFlags() {
	if c.RCO {
		c.SetRCOFFlags()
	}

	if c.Renderer != "" {
		c.SetFFlagRenderer()
	}
}
