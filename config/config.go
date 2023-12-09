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
	Channel       string        `toml:"channel"`
	Launcher      string        `toml:"launcher"`
	Renderer      string        `toml:"renderer"`
	DiscordRPC    bool          `toml:"discord_rpc"`
	ForcedVersion string        `toml:"forced_version"`
	Dxvk          bool          `toml:"dxvk"`
	FFlags        roblox.FFlags `toml:"fflags"`
	Env           Environment   `toml:"env"`
	ForcedGpu     string        `toml:"gpu"`
	GameMode      bool          `toml:"gamemode"`
}

type Config struct {
	WineRoot          string      `toml:"wineroot"`
	DxvkVersion       string      `toml:"dxvk_version"`
	MultipleInstances bool        `toml:"multiple_instances"`
	SanitizeEnv       bool        `toml:"sanitize_env"`
	Global            Binary      `toml:"global"`
	Player            Binary      `toml:"player"`
	Studio            Binary      `toml:"studio"`
	env               Environment `toml:"env"` // kept for compatibilty
	Splash            Splash      `toml:"splash"`
}

var (
	ErrInvalidRenderer  = errors.New("invalid renderer given")
	ErrNeedDXVKRenderer = errors.New("dxvk is only valid with d3d renderers")
	ErrWineRootAbs      = errors.New("ensure that the wine root given is an absolute path")
	ErrWineRootInvalid  = errors.New("invalid wine root given")
)

func Load(path string) (Config, error) {
	cfg := Default()

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Println("Using default configuration")

		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}

	return cfg, cfg.setup()
}

func (c *Config) globalize() error {
	// for compatibility
	if err := mergo.Merge(&c.Global.Env, c.env, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
		return err
	}

	if err := mergo.Merge(&c.Player, c.Global, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
		return err
	}

	if err := mergo.Merge(&c.Studio, c.Global, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
		return err
	}

	return nil
}

func Default() Config {
	return Config{
		DxvkVersion: "2.3",

		// Global should only be used to set strings here.
		Global: Binary{
			ForcedGpu: "prime-discrete",
			Renderer:  "D3D11",
			Dxvk:      true,
			GameMode:  true,
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
			Env: Environment{
				"OBS_VKCAPTURE": "1",
			},
		},
		Studio: Binary{},

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
		return ErrInvalidRenderer
	}

	if !strings.HasPrefix(b.Renderer, "D3D11") && b.Dxvk {
		return ErrNeedDXVKRenderer
	}

	if err := b.FFlags.SetRenderer(b.Renderer); err != nil {
		return err
	}

	return b.pickCard()
}

func (c *Config) setup() error {
	if err := c.globalize(); err != nil {
		return err
	}

	if c.SanitizeEnv {
		util.SanitizeEnv()
	}

	if c.WineRoot != "" {
		bin := filepath.Join(c.WineRoot, "bin")

		if !filepath.IsAbs(c.WineRoot) {
			return ErrWineRootAbs
		}

		_, err := os.Stat(filepath.Join(bin, "wine"))
		if err != nil {
			return fmt.Errorf("%w: %s", ErrWineRootInvalid, err)
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
