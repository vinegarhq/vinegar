package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/util"
)

type UI struct {
	Enabled bool   `toml:"enabled"`
	Bg      uint32 `toml:"background"`
	Fg      uint32 `toml:"foreground"`
	Red     uint32 `toml:"red"`
	Accent  uint32 `toml:"accent"`
	Gray1   uint32 `toml:"gray1"`
	Gray2   uint32 `toml:"gray2"`
}

type Application struct {
	Channel        string        `toml:"channel"`
	Launcher       string        `toml:"launcher"`
	Renderer       string        `toml:"renderer"`
	ForcedVersion  string        `toml:"forced_version"`
	AutoKillPrefix bool          `toml:"auto_kill_prefix"`
	Dxvk           bool          `toml:"dxvk"`
	FFlags         roblox.FFlags `toml:"fflags"`
	Env            Environment   `toml:"env"`
}

type Config struct {
	WineRoot          string      `toml:"wineroot"`
	DxvkVersion       string      `toml:"dxvk_version"`
	MultipleInstances bool        `toml:"multiple_instances"`
	SanitizeEnv       bool        `toml:"sanitize_env"`
	Global            Application `toml:"global"`
	Player            Application `toml:"player"`
	Studio            Application `toml:"studio"`
	Env               Environment `toml:"env"`
	UI                `toml:"ui"`
}

func Load(path string) (Config, error) {
	cfg := Default()

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Println("Using default configuration")

		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}

	if err := mergo.Merge(&cfg.Player, cfg.Global, mergo.WithAppendSlice); err != nil {
		return cfg, err
	}

	if err := mergo.Merge(&cfg.Studio, cfg.Global, mergo.WithAppendSlice); err != nil {
		return cfg, err
	}

	return cfg, cfg.setup()
}

func Default() Config {
	return Config{
		DxvkVersion: "4.3",

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

		UI: UI{
			Enabled: true,
			Bg:      0x242424,
			Fg:      0xfafafa,
			Red:     0xcc241d,
			Gray1:   0x303030,
			Gray2:   0x777777,
			Accent:  0x8fbc5e,
		},
	}
}

func (c *Config) setup() error {
	if c.SanitizeEnv {
		util.SanitizeEnv()
	}

	if c.WineRoot != "" {
		bin := filepath.Join(c.WineRoot, "bin")

		if !filepath.IsAbs(c.WineRoot) {
			return errors.New("ensure that the wine root given is an absolute path")
		}

		_, err := os.Stat(filepath.Join(bin, "wine"))
		if err != nil {
			return fmt.Errorf("invalid wine root given: %s", err)
		}

		c.Global.Env["PATH"] = bin + ":" + os.Getenv("PATH")
		os.Unsetenv("WINEDLLPATH")
		log.Printf("Using Wine Root: %s", c.WineRoot)
	}

	if !roblox.ValidRenderer(c.Player.Renderer) || !roblox.ValidRenderer(c.Studio.Renderer) {
		return fmt.Errorf("invalid renderer given to either player or studio")
	}

	c.Env.Setenv()
	c.Global.Env.Setenv()

	return nil
}
