package config

import (
	_ "embed"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/aubun/internal/dirs"
	"github.com/vinegarhq/aubun/roblox"
)

//go:embed config.toml
var DefaultConfig string

type Environment map[string]string

type Application struct {
	Channel      string        `toml:"channel"`
	Renderer     string        `toml:"renderer"`
	ForcedVerson string        `toml:"forced_version"`
	Dxvk         bool          `toml:"dxvk"`
	FFlags       roblox.FFlags `toml:"fflags"`
	Env          Environment   `toml:"env"`
}

type Config struct {
	WineRoot string        `toml:"wineroot"`
	Player   Application   `toml:"player"`
	Studio   Application   `toml:"studio"`
	FFlags   roblox.FFlags `toml:"fflags"`
	Env      Environment   `toml:"env"`
}

func Load() Config {
	var cfg Config

	path := filepath.Join(dirs.Config, "config.toml")

	// This is a default hard-coded configuration. It should not fail.
	_, _ = toml.Decode(DefaultConfig, &cfg)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Println("Using default configuration")

		return cfg
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		log.Printf("Failed to load configuration: %s, using default configuration", err)
	}

	if cfg.WineRoot != "" {
		log.Printf("Using Wine Root: %s", cfg.WineRoot)
		bin := filepath.Join(cfg.WineRoot, "bin")

		if !filepath.IsAbs(cfg.WineRoot) {
			log.Fatal("ensure that the wine root given is an absolute path")
		}

		_, err := os.Stat(filepath.Join(bin, "wine"))
		if err != nil {
			log.Fatalf("invalid wine root given: %s", err)
		}

		cfg.Env["PATH"] = bin + ":" + os.Getenv("PATH")
	}

	return cfg
}

func (e *Environment) Setenv() {
	for name, value := range *e {
		os.Setenv(name, value)
	}
}

func (c *Config) Print() error {
	return toml.NewEncoder(os.Stdout).Encode(c)
}
