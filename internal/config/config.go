package config

import (
	_ "embed"
	"log"
	"os"
	"errors"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/aubun/internal/dirs"
)

//go:embed config.toml
var DefaultConfig string

var Renderers = []string{
	"OpenGL",
	"D3D11FL10",
	"D3D11",
	"Vulkan",
}

type (
	FFlags      map[string]interface{}
	Environment map[string]string
)

type Application struct {
	Channel  string
	Renderer string
	Dxvk     bool
	Force    bool
	FFlags
	Env Environment
}

type Config struct {
	Player Application
	Studio Application
	FFlags
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

func (c *Config) Setenv() {
	for name, value := range c.Env {
		os.Setenv(name, value)
	}
}

func (c *Config) Print() error {
	return toml.NewEncoder(os.Stdout).Encode(c)
}
