// Package config implements types and routines to configure Vinegar.
package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sewnie/rbxbin"
	"github.com/sewnie/wine"
	"github.com/sewnie/wine/dxvk"
	"github.com/vinegarhq/vinegar/internal/dirs"
)

// Studio is a representation of the deployment and behavior
// of Roblox Studio.
type Studio struct {
	GameMode      bool          `toml:"gamemode"`
	WineRoot      string        `toml:"wineroot"`
	Launcher      string        `toml:"launcher"`
	Quiet         bool          `toml:"quiet"`
	DiscordRPC    bool          `toml:"discord_rpc"`
	ForcedVersion string        `toml:"forced_version"`
	Channel       string        `toml:"channel"`
	Dxvk          bool          `toml:"dxvk"`
	DxvkVersion   string        `toml:"dxvk_version"`
	WebView       string        `toml:"webview"`
	ForcedGpu     string        `toml:"gpu"`
	Renderer      string        `toml:"renderer"`
	Env           Environment   `toml:"env"`
	FFlags        rbxbin.FFlags `toml:"fflags"`
}

// Config is a representation of the Vinegar configuration.
type Config struct {
	Studio      Studio      `toml:"studio"`
	Env         Environment `toml:"env"`
}

var (
	ErrWineRootAbs     = errors.New("wine root path is not an absolute path")
	ErrWineRootInvalid = errors.New("no wine binary present in wine root")
)

// Load will load the configuration file; if it doesn't exist, it
// will fallback to the default configuration.
func Load() (*Config, error) {
	d := Default()

	if _, err := os.Stat(dirs.ConfigPath); errors.Is(err, os.ErrNotExist) {
		return d, d.Setup()
	}

	if _, err := toml.DecodeFile(dirs.ConfigPath, &d); err != nil {
		return d, err
	}

	return d, d.Setup()
}

// Default returns a default configuration.
func Default() *Config {
	return &Config{
		Env: Environment{
			"WINEARCH":                    "win64",
			"WINEDEBUG":                   "fixme-all,err-kerberos,err-ntlm",
			"WINEESYNC":                   "1",
			"WINEDLLOVERRIDES":            "dxdiagn,winemenubuilder.exe,mscoree,mshtml=",
			"DXVK_LOG_LEVEL":              "warn",
			"DXVK_LOG_PATH":               "none",
			"MESA_GL_VERSION_OVERRIDE":    "4.4",
			"__GL_THREADED_OPTIMIZATIONS": "1",
			"VK_LOADER_LAYERS_ENABLE":     "VK_LAYER_VINEGAR_VinegarLayer",
		},

		Studio: Studio{
			Dxvk:        false,
			DxvkVersion: "2.5.3",
			WebView:     "",
			GameMode:    true,
			ForcedGpu:   "prime-discrete",
			Renderer:    "Vulkan",
			Channel:     "LIVE",
			DiscordRPC:  true,
			FFlags:      make(rbxbin.FFlags),
			Env:         make(Environment),
		},
	}
}

func (s *Studio) LauncherPath() (string, error) {
	return exec.LookPath(strings.Fields(s.Launcher)[0])
}

func (c *Config) Setup() error {
	if err := c.Studio.setup(); err != nil {
		return fmt.Errorf("studio: %w", err)
	}

	// Required to read Roblox logs.
	c.Env["WINEDEBUG"] += ",warn+debugstr"
	c.Env["WINEDLLOVERRIDES"] += ";" + dxvk.EnvOverride(c.Studio.Dxvk)
	c.Env.Setenv()

	return nil
}

func (s *Studio) setup() error {
	s.Env.Setenv()

	if s.Launcher != "" {
		if _, err := s.LauncherPath(); err != nil {
			return fmt.Errorf("bad launcher: %w", err)
		}
	}

	pfx := wine.New("", s.WineRoot)

	if s.WineRoot != "" {
		w := pfx.Wine("")
		if w.Err != nil {
			return errors.New("invalid wineroot")
		}
	}

	if pfx.IsProton() {
		// https://github.com/bottlesdevs/Bottles/issues/3485
		s.Dxvk = true
	}

	if err := s.FFlags.SetRenderer(s.Renderer); err != nil {
		return err
	}

	if s.Channel == "LIVE" || s.Channel == "live" {
		s.Channel = ""
	}

	return s.pickCard()
}
