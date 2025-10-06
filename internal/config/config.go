// Package config implements types and routines to configure Vinegar.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sewnie/rbxbin"
	"github.com/sewnie/wine"
	"github.com/sewnie/wine/dxvk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
)

type Studio struct {
	GameMode bool `toml:"gamemode" group:"Performance" row:"Apply system optimizations. May improve performance."`

	ForcedGpu   string `toml:"gpu" group:"Rendering" row:"Named or Indexed GPU (ex. integrated, prime-discrete, 0)"`
	DXVK        bool   `toml:"dxvk" group:"Rendering" row:"Improve Direct3D compatibility by translating it to Vulkan"`
	DXVKVersion string `toml:"dxvk_version" group:"Rendering" row:"DXVK Version"`
	Renderer    string `toml:"renderer" group:"Rendering" row:"D3D11 for DXVK,vals,D3D11,D3D11FL10,Vulkan,OpenGL"` // Enum reflection is impossible

	WineRoot string `toml:"wineroot" group:"Custom Wine" row:"Installation Directory,path"`
	WebView  string `toml:"webview" group:"Custom Wine" row:"WebView2 Runtime Version"`
	Launcher string `toml:"launcher" group:"Custom Wine" row:"Launcher Command"`

	ForcedVersion string `toml:"forced_version" group:"Deployment Overrides" row:"Studio Deployment Version"`
	Channel       string `toml:"channel" group:"Deployment Overrides" row:"Studio Update Channel"`

	Quiet      bool          `toml:"quiet" group:"Behavior" row:"Prevent capturing of Studio's debugging logs"`
	DiscordRPC bool          `toml:"discord_rpc" group:"Behavior" row:"Display your development status on your Discord profile"`
	Env        Environment   `toml:"env" group:"Studio Environment Variables"`
	FFlags     rbxbin.FFlags `toml:"fflags" group:"Studio Fast Flags"`
}

type Config struct {
	Debug  bool        `toml:"debug" group:"Behavior" row:"Show API requests made and disable safety checks"`
	Studio Studio      `toml:"studio"`
	Env    Environment `toml:"env" group:"Environment"`
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
		Debug: false,

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
			DXVK:        false,
			DXVKVersion: "2.7",
			WebView:     "109.0.1518.140", // Last known win7
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
	level := slog.LevelInfo
	if c.Debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(
		logging.NewHandler(os.Stderr, level)))

	if !c.Debug && os.Getenv("WAYLAND_DISPLAY") != "" {
		c.Env["DISPLAY"] = ""
	}

	c.Env["WINEDEBUG"] += ",warn+debugstr" // required to read Roblox logs
	c.Env["WINEDLLOVERRIDES"] += ";" + dxvk.EnvOverride(c.Studio.DXVK)
	c.Env.Setenv()

	if err := c.Studio.setup(); err != nil {
		return fmt.Errorf("studio: %w", err)
	}

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
		s.DXVK = true
	}

	if err := s.FFlags.SetRenderer(s.Renderer); err != nil {
		return err
	}

	if s.Channel == "LIVE" || s.Channel == "live" {
		s.Channel = ""
	}

	return s.pickCard()
}
