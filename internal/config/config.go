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
	Channel  string
	Renderer string
	Dxvk     bool
	roblox.FFlags
	Env Environment
}

type Config struct {
	Player Application
	Studio Application
	roblox.FFlags
	Env Environment
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

	return cfg
}

func (e *Environment) Setenv() {
	for name, value := range (*e) {
		os.Setenv(name, value)
	}
}

func (c *Config) Print() error {
	return toml.NewEncoder(os.Stdout).Encode(c)
}
