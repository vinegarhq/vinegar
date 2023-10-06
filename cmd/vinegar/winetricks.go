package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/wine"
)

func find_winetricks(wineroot string) (string, error) {
	path, err := exec.LookPath(filepath.Join(wineroot, "bin", "winetricks"))
	if err != nil {
		log.Printf("winetricks not found in %s, trying PATH", wineroot)
		path, err = exec.LookPath("winetricks")
		if err != nil {
			return "", err
		}
	}
	return path, nil
}

func LaunchWinetricks(pfx *wine.Prefix, cfg *config.Config) error {
	log.Println("Launching winetricks")
	path, err := find_winetricks(cfg.WineRoot)
	if err != nil {
		log.Println("Failed to find winetricks:", err)
		return err
	}
	log.Printf("Found winetricks at %s\n", path)

	envVars := []string{
		"WINEPREFIX=" + pfx.Dir,
	}

	if cfg.WineRoot != "" {
		envVars = append(envVars, "WINE="+filepath.Join(cfg.WineRoot, "bin", "wine"))
	}

	cmd := exec.Command("winetricks")
	cmd.Env = append(os.Environ(), envVars...)

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}

	log.Printf("Output: %s", output)

	return nil
}
