package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"dario.cat/mergo"
	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/util"
)

type Splash struct {
	Enabled bool   `toml:"enabled"`
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
	Env               Environment `toml:"env"`
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

		Global: Binary{
			ForcedGpu: "prime-discrete",
		},
		Player: Binary{
			Dxvk:           true,
			AutoKillPrefix: true,
			FFlags: roblox.FFlags{
				"DFIntTaskSchedulerTargetFps": 640,
			},
		},
		Studio: Binary{
			Dxvk:           true,
			AutoKillPrefix: true,
		},

		Splash: Splash{
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

func ParseBinary(b Binary, kind string) error {
	if !roblox.ValidRenderer(b.Renderer) {
		return fmt.Errorf("invalid renderer given to " + kind)
	}

	//Validate and sanitize ForcedGpu
	switch b.ForcedGpu {
	case "":
	case "integrated":
	case "prime-discrete":
	default:
		//Sanitize value so it's case insensitive and doesn't care about "0x".
		b.ForcedGpu = strings.ReplaceAll(strings.ToLower(b.ForcedGpu), "0x", "")
		if strings.Contains(b.ForcedGpu, ":") { //Interpret as card vid:nid; do nothing
		} else { //Interpret as index.
			_, err := strconv.Atoi(b.ForcedGpu)
			if err != nil {
				return errors.New("invalid gpu for " + kind + ", it must be \"integrated\", \"prime-discrete\", the card's vid:nid or its index")
			}
		}
	}

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

		c.Env["PATH"] = bin + ":" + os.Getenv("PATH")
		os.Unsetenv("WINEDLLPATH")
		log.Printf("Using Wine Root: %s", c.WineRoot)
	}

	//Parse global first to avoid showing errors from settings which player and studio inherited from global.
	err := ParseBinary(c.Global, "global")
	if err != nil {
		return err
	}
	err = errors.Join(
		ParseBinary(c.Player, "player"),
		ParseBinary(c.Studio, "studio"),
	)
	if err != nil {
		return err
	}

	c.Env.Setenv()

	return nil
}
