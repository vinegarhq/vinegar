package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dario.cat/mergo"
	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/util"
)

type Splash struct {
	Enabled bool   `toml:"enabled"`
	Style   string `toml:"style"`
	Bg      uint32 `toml:"background"`
	Fg      uint32 `toml:"foreground"`
	Red     uint32 `toml:"red"`
	Accent  uint32 `toml:"accent"`
	Gray1   uint32 `toml:"gray1"`
	Gray2   uint32 `toml:"gray2"`
}

type Binary struct {
	Channel        string        `toml:"channel"`
	Launcher       string        `toml:"launcher"`
	Renderer       string        `toml:"renderer"`
	DiscordRPC     bool          `toml:"discord_rpc"`
	ForcedVersion  string        `toml:"forced_version"`
	AutoKillPrefix bool          `toml:"auto_kill_prefix"`
	Dxvk           bool          `toml:"dxvk"`
	FFlags         roblox.FFlags `toml:"fflags"`
	Env            Environment   `toml:"env"`
	ForcedGpu      string        `toml:"gpu"`
}

type Config struct {
	WineRoot          string      `toml:"wineroot"`
	DxvkVersion       string      `toml:"dxvk_version"`
	MultipleInstances bool        `toml:"multiple_instances"`
	SanitizeEnv       bool        `toml:"sanitize_env"`
	Global            Binary      `toml:"global"`
	Player            Binary      `toml:"player"`
	Studio            Binary      `toml:"studio"`
	Env               Environment `toml:"env"` // kept for compatibilty
	Splash            Splash      `toml:"splash"`
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

	// Compatibility
	if err := mergo.Merge(&cfg.Global.Env, cfg.Env, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
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
		DxvkVersion: "2.3",

		Global: Binary{
			ForcedGpu:      "prime-discrete",
			Dxvk:           true,
			AutoKillPrefix: true,
			Env: Environment{
				"WINEARCH":         "win64",
				"WINEDEBUG":        "err-kerberos,err-ntlm",
				"WINEESYNC":        "1",
				"WINEDLLOVERRIDES": "dxdiagn,winemenubuilder.exe,mscoree,mshtml=",

				"DXVK_LOG_LEVEL": "warn",
				"DXVK_LOG_PATH":  "none",

				"MESA_GL_VERSION_OVERRIDE":    "4.4",
				"__GL_THREADED_OPTIMIZATIONS": "1",
			},
		},
		Player: Binary{
			DiscordRPC: true,
			FFlags: roblox.FFlags{
				"DFIntTaskSchedulerTargetFps": 640,
			},
		},

		Splash: Splash{
			Enabled: true,
			Bg:      0x242424,
			Fg:      0xfafafa,
			Red:     0xbc3c3c,
			Gray1:   0x303030,
			Gray2:   0x777777,
			Accent:  0x8fbc5e,
		},
	}
}

func (b *Binary) setup() error {
	if !roblox.ValidRenderer(b.Renderer) {
		return errors.New("invalid renderer given")
	}

	if err := b.pickCard(); err != nil {
		return err
	}

	if !strings.HasPrefix(b.Renderer, "D3D11") && b.Dxvk {
		return errors.New("dxvk is only valid with d3d renderers")
	}

	b.Env.Setenv()
	return nil
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

	if err := c.Player.setup(); err != nil {
		return fmt.Errorf("player: %w", err)
	}

	if err := c.Studio.setup(); err != nil {
		return fmt.Errorf("studio: %w", err)
	}

	c.Global.Env.Setenv()

	return nil
}
