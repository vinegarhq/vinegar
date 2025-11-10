// Package config implements types and routines to configure Vinegar.
package config

import (
	"errors"
	"fmt"
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
)

// Backwards compatibility to allow:
// 'dxvk = true' and move to 'dxvk = [version]'
type DxvkVersion string

func (v *DxvkVersion) UnmarshalTOML(data interface{}) error {
	switch d := data.(type) {
	case bool:
		*v = ""
		if d {
			*v = "2.7.1"
		}
	case string:
		*v = DxvkVersion(d)
	default:
		return fmt.Errorf("unsupported type: %T", d)
	}
	return nil
}

func (v DxvkVersion) String() string {
	return string(v)
}

type DesktopConfig struct {
	Name string            `toml:"name"`
	Path string            `toml:"path"`
	Env  map[string]string `toml:"env"`
}

type DesktopManager struct {
	Enabled        bool            `toml:"enabled" group:"Desktop Management" row:"Enable multiple desktop support" title:"Multi-Desktop Support"`
	AutoAssign     bool            `toml:"auto_assign" group:"Desktop Management" row:"Automatically assign desktops to instances"`
	MaxDesktops    int             `toml:"max_desktops" group:"Desktop Management" row:"Maximum number of desktops,entry,Number,10"`
	DefaultDesktop string          `toml:"default_desktop" group:"Desktop Management" row:"Default desktop name,entry,Name,desktop-1"`
	DesktopPrefix  string          `toml:"desktop_prefix" group:"Desktop Management" row:"Prefix for desktop names,entry,Prefix,desktop-"`
	Desktops       []DesktopConfig `toml:"desktops" group:"Desktop Management"`
	IsolationLevel string          `toml:"isolation_level" group:"Desktop Management" row:"Isolation level,vals,full,basic,minimal"`
}

type Studio struct {
	WebView  string `toml:"webview" group:"" row:"Disable if nonfunctional,entry,WebView2 Version,141.0.3537.71" title:"Web Pages"`
	WineRoot string `toml:"wineroot" group:"" row:"Installation Directory,path"`
	Launcher string `toml:"launcher" group:"" row:"Launcher Command (ex. gamescope)"`

	DXVK      DxvkVersion `toml:"dxvk" group:"Rendering" row:"Improve D3D11 compatibility by translating it to Vulkan,entry,Version,2.7.1"`
	Renderer  string      `toml:"renderer" group:"Rendering" row:"Studio's Graphics Mode,vals,D3D11,D3D11FL10,Vulkan,OpenGL"` // Enum reflection is impossible
	ForcedGPU string      `toml:"gpu" group:"Rendering" row:"Named or Indexed GPU (ex. integrated or 0)"`

	DiscordRPC bool `toml:"discord_rpc" group:"Behavior" row:"Display your development status on your Discord profile" title:"Share Activity on Discord"`
	GameMode   bool `toml:"gamemode" group:"Behavior" row:"Apply system optimizations. May improve performance."`

	Env    map[string]string `toml:"env" group:"Environment"`
	FFlags rbxbin.FFlags     `toml:"fflags" group:"Fast Flags"`

	ForcedVersion string `toml:"forced_version" group:"Deployment Overrides" row:"Studio Deployment Version"`
	Channel       string `toml:"channel" group:"Deployment Overrides" row:"Studio Update Channel"`
}

type Config struct {
	Studio         Studio         `toml:"studio"`
	DesktopManager DesktopManager `toml:"desktop_manager"`
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
			DXVK:       "",
			WebView:    "141.0.3537.71", // YMMV
			GameMode:   true,
			ForcedGPU:  "prime-discrete",
			Renderer:   "D3D11",
			Channel:    "",
			DiscordRPC: true,
			FFlags:     make(rbxbin.FFlags),
			Env: map[string]string{
				"WINEESYNC": "1",
			},

			DesktopManager: DesktopManager{
				Enabled:        false,
				AutoAssign:     true,
				MaxDesktops:    10,
				DefaultDesktop: "desktop-1",
				DesktopPrefix:  "desktop-",
				Desktops:       []DesktopConfig{},
				IsolationLevel: "basic",
			},
		},
	}
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

	card, err := c.Studio.card()
	if err != nil {
		return nil, err
	}
	if card != nil {
		env["MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE"] = "1"
		env["DRI_PRIME"] = "pci-" + strings.NewReplacer(":", "_", ".", "_").
			Replace(path.Base(card.Device))

		env["__GLX_VENDOR_LIBRARY_NAME"] = "mesa"
		if strings.HasPrefix(card.Driver, "nvidia") {
			env["__GLX_VENDOR_LIBRARY_NAME"] = "nvidia"
		}
	}

	env["WINEDEBUG"] += ",warn+debugstr" // required to read Roblox logs
	env["XR_LOADER_DEBUG"] = "none"      // already shown in Roblox log
	env["WINEDLLOVERRIDES"] += ";" + "dxdiagn,winemenubuilder.exe,mscoree,mshtml="
	if !c.Debug {
		env["WINEDEBUG"] += ",fixme-all,err-kerberos,err-ntlm,err-combase"
	}

	if c.Studio.DXVK != "" {
		if !c.Debug {
			env["DXVK_LOG_LEVEL"] = "warn"
		}
		env["DXVK_LOG_PATH"] = "none"
	}
	env["VK_LOADER_LAYERS_ENABLE"] = "VK_LAYER_VINEGAR_VinegarLayer"
	env["MESA_GL_VERSION_OVERRIDE"] = "4.4"
	env["__GL_THREADED_OPTIMIZATIONS"] = "1"

	for k, v := range env {
		pfx.Env = append(pfx.Env, k+"="+v)
	}

	dxvk.EnvOverride(pfx, c.Studio.DXVK != "")

	slog.Debug("Using Prefix environment", "env", pfx.Env)

	if c.Studio.WineRoot != "" {
		w := pfx.Wine("")
		if w.Err != nil {
			return nil, w.Err
		}
	}

	return pfx, nil
}

func (c *Config) PrefixForDesktop(desktopName string) (*wine.Prefix, error) {
	var prefixPath string

	if c.DesktopManager.Enabled && desktopName != "" {
		prefixPath = path.Join(dirs.Prefixes, desktopName)
	} else {
		// Backward compatibility: use single "studio" prefix
		prefixPath = path.Join(dirs.Prefixes, "studio")
	}

	pfx := wine.New(prefixPath, c.Studio.WineRoot)

	env := maps.Clone(c.Studio.Env)

	// Add desktop-specific environment variables if multi-desktop is enabled
	if c.DesktopManager.Enabled && desktopName != "" {
		// Find desktop config and merge its environment
		for _, desktop := range c.DesktopManager.Desktops {
			if desktop.Name == desktopName {
				maps.Copy(env, desktop.Env)
				break
			}
		}

		// Add desktop identification
		env["VINEGAR_DESKTOP"] = desktopName
		env["VINEGAR_DESKTOP_ID"] = strings.TrimPrefix(desktopName, c.DesktopManager.DesktopPrefix)
	}

	card, err := c.Studio.card()
	if err != nil {
		return nil, err
	}
	if card != nil {
		env["MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE"] = "1"
		env["DRI_PRIME"] = "pci-" + strings.NewReplacer(":", "_", ".", "_").
			Replace(path.Base(card.Device))

		env["__GLX_VENDOR_LIBRARY_NAME"] = "mesa"
		if strings.HasPrefix(card.Driver, "nvidia") {
			env["__GLX_VENDOR_LIBRARY_NAME"] = "nvidia"
		}
	}

	env["WINEDEBUG"] += ",warn+debugstr" // required to read Roblox logs
	env["XR_LOADER_DEBUG"] = "none"      // already shown in Roblox log
	env["WINEDLLOVERRIDES"] += ";" + "dxdiagn,winemenubuilder.exe,mscoree,mshtml="
	if !c.Debug {
		env["WINEDEBUG"] += ",fixme-all,err-kerberos,err-ntlm,err-combase"
	}

	if c.Studio.DXVK != "" {
		if !c.Debug {
			env["DXVK_LOG_LEVEL"] = "warn"
		}
		env["DXVK_LOG_PATH"] = "none"
	}
	env["VK_LOADER_LAYERS_ENABLE"] = "VK_LAYER_VINEGAR_VinegarLayer"
	env["MESA_GL_VERSION_OVERRIDE"] = "4.4"
	env["__GL_THREADED_OPTIMIZATIONS"] = "1"

	for k, v := range env {
		pfx.Env = append(pfx.Env, k+"="+v)
	}

	dxvk.EnvOverride(pfx, c.Studio.DXVK != "")

	slog.Debug("Using Prefix environment", "env", pfx.Env, "desktop", desktopName)

	if c.Studio.WineRoot != "" {
		w := pfx.Wine("")
		if w.Err != nil {
			return nil, w.Err
		}
	}

	return pfx, nil
}

// GetAvailableDesktops returns a list of available desktop names
func (c *Config) GetAvailableDesktops() []string {
	if !c.DesktopManager.Enabled {
		return []string{"studio"}
	}

	var desktops []string
	for _, desktop := range c.DesktopManager.Desktops {
		desktops = append(desktops, desktop.Name)
	}

	// If no desktops configured, return default
	if len(desktops) == 0 {
		desktops = append(desktops, c.DesktopManager.DefaultDesktop)
	}

	return desktops
}

// AssignDesktop assigns a desktop for a new instance
func (c *Config) AssignDesktop() string {
	if !c.DesktopManager.Enabled || !c.DesktopManager.AutoAssign {
		return c.DesktopManager.DefaultDesktop
	}

	// Simple round-robin assignment for now
	// In a more sophisticated implementation, we could track active instances
	desktops := c.GetAvailableDesktops()
	if len(desktops) == 0 {
		return c.DesktopManager.DefaultDesktop
	}

	// For now, return the first available desktop
	// TODO: Implement proper load balancing
	return desktops[0]
}
