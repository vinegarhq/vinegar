// Package config implements types and routines to configure Vinegar.
package config

import (
	"bytes"
	"errors"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sewnie/rbxbin"
	"github.com/sewnie/wine"
	"github.com/sewnie/wine/dxvk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/sysinfo"
)

const (
	dxvkVersion      = "2.7.1"
	dxvkSarekVersion = "Sarek-1.11.0-async"
	webViewVersion   = "143.0.3650.96"
)

type Studio struct {
	WebView  WebViewOption `toml:"webview" group:"" row:"Disable if nonfunctional,WebView2 Version" title:"Web Pages"`
	WineRoot string        `toml:"wineroot" group:"" row:"Wine Installation,path"`
	Launcher string        `toml:"launcher" group:"" row:"Launcher Command (ex. gamescope)"`
	Desktop  DesktopOption `toml:"virtual_desktop" group:"" row:"Create an isolated window for each Studio instance,Window resolution (eg. 1920x1080)" title:"Virtual Desktops"`

	Renderer Renderer `toml:"renderer" group:"Rendering" row:"Studio's Graphics Mode"` // Enum reflection is impossible

	DiscordRPC bool `toml:"discord_rpc" group:"Behavior" row:"Display your development status on your Discord profile" title:"Share Activity on Discord"`
	GameMode   bool `toml:"gamemode" group:"Behavior" row:"Apply system optimizations. May improve performance."`

	Env    map[string]string `toml:"env" group:"Environment"`
	FFlags rbxbin.FFlags     `toml:"fflags" group:"Fast Flags"`

	ForcedVersion string `toml:"forced_version" group:"Deployment Overrides" row:"Studio Deployment Version"`
	Channel       string `toml:"channel" group:"Deployment Overrides" row:"Studio Update Channel"`
}

type Config struct {
	Studio Studio `toml:"studio"`
	// Only adds to Studio.Env, reserved for backwards compatibility
	Env   map[string]string `toml:"env" group:"hidden"`
	Debug bool              `toml:"debug" group:"Behavior" row:"Output Studio logs and Web API requests"`
}

var (
	ErrWineRootAbs     = errors.New("wine root path is not an absolute path")
	ErrWineRootInvalid = errors.New("no wine binary present in wine root")
)

// Load will load the configuration file; if it doesn't exist, it
// will fallback to the default configuration.
func Load() (*Config, error) {
	cfg := Default()

	if _, err := os.Stat(dirs.ConfigPath); errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(dirs.ConfigPath, &cfg); err != nil {
		return cfg, err
	}

	maps.Copy(cfg.Studio.Env, cfg.Env)
	cfg.Env = nil

	logging.LoggerLevel = slog.LevelInfo
	if cfg.Debug {
		logging.LoggerLevel = slog.LevelDebug
	}

	return cfg, nil
}

// Default returns a default configuration.
func Default() *Config {
	return &Config{
		Debug: false,

		Env: make(map[string]string),

		Studio: Studio{
			WineRoot:   dirs.WinePath,
			WebView:    WebViewOption(webViewVersion),
			GameMode:   true,
			Renderer:   "D3D11",
			Channel:    "",
			DiscordRPC: true,
			FFlags:     make(rbxbin.FFlags),
			Env:        make(map[string]string),
		},
	}
}

func (s *Studio) UnmarshalTOML(data interface{}) error {
	// prevent recursion by typing
	type Alias Studio
	proxy := struct {
		*Alias
		DXVK string `toml:"dxvk"`
	}{Alias: (*Alias)(s)}

	// encode to and back to retrieve all options from raw TOML
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}
	if _, err := toml.Decode(buf.String(), &proxy); err != nil {
		return err
	}

	s = (*Studio)(proxy.Alias)

	if proxy.DXVK != "" {
		slog.Warn("The DXVK option alongside it's versioning has been deprecated, setting Renderer")
		s.Renderer = "DXVK"
	}
	return nil
}

func (s *Studio) LauncherPath() (string, error) {
	return exec.LookPath(strings.Fields(s.Launcher)[0])
}

func (c *Config) Prefix() (*wine.Prefix, error) {
	pfx := wine.New(
		path.Join(dirs.Prefixes, "studio"),
		c.Studio.WineRoot,
	)

	env := maps.Clone(c.Studio.Env)

	if len(sysinfo.Cards) > 1 {
		env["MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE"] = "1"
	}

	env["WINEDEBUG"] += ",warn+debugstr" // required to read Roblox logs
	env["XR_LOADER_DEBUG"] = "none"      // already shown in Roblox log
	env["WINEDLLOVERRIDES"] += ";" + "dxdiagn,winemenubuilder.exe,mscoree,mshtml="
	if !c.Debug {
		env["WINEDEBUG"] += ",fixme-all,err-kerberos,err-ntlm,err-combase"
	}

	env["WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS"] = "--in-process-gpu "

	switch c.Studio.Renderer {
	case "D3D11", "OpenGL":
		env["WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS"] += "--use-angle=gl"
	default: // all other options are vulkan-esque
		env["WINE_D3D_CONFIG"] = "renderer=vulkan"
		env["WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS"] += "--use-angle=d3d11"
	}

	if c.Studio.Renderer.IsDXVK() {
		if !c.Debug {
			env["DXVK_LOG_LEVEL"] = "warn"
		}
		env["DXVK_LOG_PATH"] = "none"
		env["DXVK_STATE_CACHE_PATH"] = dirs.Cache
	}
	env["VK_LOADER_LAYERS_ENABLE"] = "VK_LAYER_VINEGAR_VinegarLayer"

	for k, v := range env {
		pfx.Env = append(pfx.Env, k+"="+v)
	}

	dxvk.EnvOverride(pfx, c.Studio.Renderer.IsDXVK())

	slog.Debug("Using Prefix environment", "env", pfx.Env)

	return pfx, nil
}
