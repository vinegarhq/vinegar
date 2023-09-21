package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/util"
)

var Path = filepath.Join(dirs.Config, "config.toml")

type Environment map[string]string

type Application struct {
	Channel        string        `toml:"channel"`
	Renderer       string        `toml:"renderer"`
	ForcedVersion  string        `toml:"forced_version"`
	AutoKillPrefix bool          `toml:"auto_kill_prefix"`
	Dxvk           bool          `toml:"dxvk"`
	FFlags         roblox.FFlags `toml:"fflags"`
	Env            Environment   `toml:"env"`
}

type Config struct {
	Launcher          string      `toml:"launcher"`
	WineRoot          string      `toml:"wineroot"`
	DxvkVersion       string      `toml:"dxvk_version"`
	MultipleInstances bool        `toml:"multiple_instances"`
	SanitizeEnv       bool        `toml:"sanitize_env"`
	Player            Application `toml:"player"`
	Studio            Application `toml:"studio"`
	Env               Environment `toml:"env"`
}

func Load() (Config, error) {
	cfg := Default()

	if _, err := os.Stat(Path); errors.Is(err, os.ErrNotExist) {
		log.Println("Using default configuration")

		return cfg, nil
	}

	if _, err := toml.DecodeFile(Path, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode configuration file: %w", err)
	}

	if err := cfg.Setup(); err != nil {
		return cfg, fmt.Errorf("failed to setup configuration: %w", err)
	}

	return cfg, nil
}

func Default() Config {
	return Config{
		DxvkVersion: "2.3",

		Env: Environment{
			"WINEARCH":         "win64",
			"WINEDEBUG":        "err-kerberos,err-ntlm",
			"WINEESYNC":        "1",
			"WINEDLLOVERRIDES": "dxdiagn=d;winemenubuilder.exe=d",

			"DXVK_LOG_LEVEL": "warn",
			"DXVK_LOG_PATH":  "none",

			"MESA_GL_VERSION_OVERRIDE":    "4.4",
			"__GL_THREADED_OPTIMIZATIONS": "1",
		},

		Player: Application{
			Dxvk:           true,
			AutoKillPrefix: true,
			FFlags: roblox.FFlags{
				"DFIntTaskSchedulerTargetFps": 640,
			},
		},
		Studio: Application{
			Dxvk: true,
		},
	}
}

func (e *Environment) Setenv() {
	for name, value := range *e {
		os.Setenv(name, value)
	}
}

func (c *Config) Setup() error {
	if c.SanitizeEnv {
		util.SanitizeEnv()
	}

	if c.WineRoot != "" {
		bin := filepath.Join(c.WineRoot, "bin")

		if !filepath.IsAbs(c.WineRoot) {
			return errors.New("ensure that the wine root given is an absolute path")
		}

		wine_path := filepath.Join(bin, "wine")
		wine64_path := filepath.Join(bin, "wine64")
		_, err := os.Stat(wine_path)
		_, err2 := os.Stat(wine64_path)

		if err != nil && err2 == nil {
			//Workaround for 64bit-only wine builds; they don't actually have a wine binary, so make a hard link to it instead.
			//Technically we could just infer this during runtime, but making the hard link is the easiest solution.
			//If we don't make a link, system 32bit binaries might be used along with WineRoot's 64bit binaries... (bad)
			err := os.Symlink(wine64_path, wine_path)

			if err != nil {
				return fmt.Errorf(`error while creating the hard link "%s" to "%s": %s.
				if vinegar has no permissions to create a link, try doing it manually`, wine_path, wine64_path, err)
			}
		} else {
			return fmt.Errorf("invalid wine root given: %s | %s", err, err2)
		}

		c.Env["PATH"] = bin + ":" + os.Getenv("PATH")
		os.Unsetenv("WINEDLLPATH")
		log.Printf("Using Wine Root: %s", c.WineRoot)
	}

	if !roblox.ValidRenderer(c.Player.Renderer) || !roblox.ValidRenderer(c.Studio.Renderer) {
		return fmt.Errorf("invalid renderer given to either player or studio")
	}

	c.Env.Setenv()

	return nil
}
