package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/nxadm/tail"
	bsrpc "github.com/vinegarhq/vinegar/bloxstraprpc"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/config/state"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/splash"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/roblox/version"
	"github.com/vinegarhq/vinegar/util"
	"github.com/vinegarhq/vinegar/wine"
	"github.com/vinegarhq/vinegar/wine/dxvk"
)

type Binary struct {
	Splash *splash.Splash

	GlobalConfig *config.Config
	Config       *config.Binary

	Alias   string
	Name    string
	Dir     string
	Prefix  *wine.Prefix
	Type    roblox.BinaryType
	Version version.Version
	Started time.Time
}

func NewBinary(bt roblox.BinaryType, cfg *config.Config, pfx *wine.Prefix) Binary {
	var bcfg config.Binary

	switch bt {
	case roblox.Player:
		bcfg = cfg.Player
	case roblox.Studio:
		bcfg = cfg.Studio
	}

	return Binary{
		Splash: splash.New(&cfg.Splash),

		GlobalConfig: cfg,
		Config:       &bcfg,

		Alias:  bt.String(),
		Name:   bt.BinaryName(),
		Type:   bt,
		Prefix: pfx,
	}
}

func (b *Binary) Run(args ...string) error {
	cmd, err := b.Command(args...)
	if err != nil {
		return err
	}

	log.Printf("Launching %s", b.Name)
	b.Splash.Message("Launching " + b.Alias)


	// Launches into foreground
	b.Started = time.Now()
	if err := cmd.Start(); err != nil {
		return err
	}

	// act as the signal holder, as roblox/wine will not do anything
	// with the INT signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		// This way, cmd.Wait() will return and the wineprefix killer
		// will be ran.
		log.Println("Killing Roblox")
		cmd.Process.Kill()
		signal.Stop(c)
	}()

	if b.Config.DiscordRPC {
		err := bsrpc.Login()
		if err != nil {
			log.Printf("Failed to authenticate Discord RPC: %s, disabling RPC", err)
			b.Config.DiscordRPC = false
		}
		// NOTE: This will panic if logout fails
		defer bsrpc.Logout()
	}

	// after FindLog() fails to find a log, assume Roblox hasn't started.
	// early. if it did, assume failure and jump to cmd.Wait(), which will give the
	// returned error to the splash screen and output if wine returns one.
	p, err := b.FindLog()
	if err != nil {
		log.Printf("%s, assuming roblox failure", err)
	} else {
		b.Splash.Close()

		go func() {
			rblxExited, err := b.TailLog(p)
			if err != nil {
				log.Printf("tail roblox log file: %s", err)
				return
			}

			if rblxExited {
				log.Println("Got Roblox shutdown")
				// give roblox two seconds to cleanup its garbage
				time.Sleep(2 * time.Second)
				// force kill the process, causing cmd.Wait() to immediately return.
				log.Println("Killing Roblox")
				cmd.Process.Kill()
				cmd.Process.Kill()
				cmd.Process.Kill()
				cmd.Process.Kill()
				cmd.Process.Kill()
				cmd.Process.Kill()
				cmd.Process.Kill()
			}
		}()
	}

	defer func() {
		// If roblox is already running, don't kill wineprefix, even if
		// auto kill prefix is enabled
		if util.CommFound("Roblox") {
			log.Println("Roblox is already running, not killing wineprefix after exit")
			return
		}

		if b.Config.AutoKillPrefix {
			b.Prefix.Kill()
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("roblox process: %w", err)
	}

	return nil
}

func (b *Binary) FindLog() (string, error) {
	appData, err := b.Prefix.AppDataDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(appData, "Local", "Roblox", "logs")
	// May not exist if roblox has its first run
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	log.Println("Polling for Roblox log file, 10 retries")
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)

		name, err := util.FindTimeFile(dir, &b.Started)
		if err == nil {
			log.Printf("Found Roblox log file: %s", name)
			return name, nil
		}
	}

	return "", fmt.Errorf("could not find roblox log file after time %s", b.Started)
}

// Boolean returned is if Roblox had exited, detecting via logs
func (b *Binary) TailLog(name string) (bool, error) {
	var a bsrpc.Activity
	const title = "WebView/InternalBrowser is broken"
	auth := false

	t, err := tail.TailFile(name, tail.Config{Follow: true, MustExist: true})
	if err != nil {
		return false, err
	}

	for line := range t.Lines {
		fmt.Fprintln(b.Prefix.Output, line.Text)

		// Easy way to figure out we are authenticated, to make a more
		// babysit message to tell the user to use quick login
		if strings.Contains(line.Text, "DID_LOG_IN") {
			auth = true
		}

		if strings.Contains(line.Text, "the local did not install any WebView2 runtime") {
			if auth {
				b.Splash.Dialog(title, "use the browser for whatever you were doing just now.")
			} else {
				b.Splash.Dialog(title, "Use Quick Log In to authenticate ('Log In With Another Device' button)")
			}
		}

		// Best we've got to know if roblox had actually quit
		if strings.Contains(line.Text, "[FLog::SingleSurfaceApp] shutDown:") {
			return true, nil
		}

		if b.Config.DiscordRPC {
			if err := a.HandleLog(line.Text); err != nil {
				log.Printf("Failed to handle Discord RPC: %s", err)
			}
		}
	}

	// this is should be unreachable
	return false, nil
}

func (b *Binary) FetchVersion() (version.Version, error) {
	b.Splash.Message("Fetching " + b.Alias)

	if b.Config.ForcedVersion != "" {
		log.Printf("WARNING: using forced version: %s", b.Config.ForcedVersion)

		return version.New(b.Type, b.Config.Channel, b.Config.ForcedVersion), nil
	}

	return version.Fetch(b.Type, b.Config.Channel)
}

func (b *Binary) Setup() error {
	ver, err := b.FetchVersion()
	if err != nil {
		return err
	}

	b.Splash.Desc(fmt.Sprintf("%s %s", ver.GUID, ver.Channel))
	b.Version = ver
	b.Dir = filepath.Join(dirs.Versions, ver.GUID)

	stateVer, err := state.Version(b.Type)
	if err != nil {
		log.Printf("Failed to retrieve stored %s version: %s", b.Name, err)
	}

	if stateVer != ver.GUID {
		log.Printf("Installing %s (%s -> %s)", b.Name, stateVer, ver.GUID)

		if err := b.Install(); err != nil {
			return err
		}
	} else {
		log.Printf("%s is up to date (%s)", b.Name, ver.GUID)
	}

	b.Config.Env.Setenv()

	if err := b.Config.FFlags.SetRenderer(b.Config.Renderer); err != nil {
		return err
	}

	if err := b.Config.FFlags.Apply(b.Dir); err != nil {
		return err
	}

	if err := dirs.OverlayDir(b.Dir); err != nil {
		return err
	}

	if err := b.SetupDxvk(); err != nil {
		return err
	}

	b.Splash.Progress(1.0)
	return nil
}

func (b *Binary) Install() error {
	b.Splash.Message("Installing " + b.Alias)

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	manifest, err := bootstrapper.FetchManifest(&b.Version)
	if err != nil {
		return err
	}

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	b.Splash.Message("Downloading " + b.Alias)
	if err := b.DownloadPackages(&manifest); err != nil {
		return err
	}

	b.Splash.Message("Extracting " + b.Alias)
	if err := b.ExtractPackages(&manifest); err != nil {
		return err
	}

	if b.Type == roblox.Studio {
		brokenFont := filepath.Join(b.Dir, "StudioFonts", "SourceSansPro-Black.ttf")

		log.Printf("Removing broken font %s", brokenFont)
		if err := os.RemoveAll(brokenFont); err != nil {
			log.Printf("Failed to remove font: %s", err)
		}
	}

	if err := bootstrapper.WriteAppSettings(b.Dir); err != nil {
		return err
	}

	if err := state.SaveManifest(&manifest); err != nil {
		return err
	}

	if err := state.CleanPackages(); err != nil {
		return err
	}

	return state.CleanVersions()
}

func (b *Binary) PerformPackages(m *bootstrapper.Manifest, fn func(bootstrapper.Package) error) error {
	donePkgs := 0
	pkgsLen := len(m.Packages)

	return m.Packages.Perform(func(pkg bootstrapper.Package) error {
		err := fn(pkg)
		if err != nil {
			return err
		}

		donePkgs++
		b.Splash.Progress(float32(donePkgs) / float32(pkgsLen))

		return nil
	})
}

func (b *Binary) DownloadPackages(m *bootstrapper.Manifest) error {
	log.Printf("Downloading %d Packages for %s", len(m.Packages), m.Version.GUID)

	return b.PerformPackages(m, func(pkg bootstrapper.Package) error {
		return pkg.Fetch(filepath.Join(dirs.Downloads, pkg.Checksum), m.DeployURL)
	})
}

func (b *Binary) ExtractPackages(m *bootstrapper.Manifest) error {
	log.Printf("Extracting %d Packages for %s", len(m.Packages), m.Version.GUID)
	pkgDirs := bootstrapper.BinaryDirectories(b.Type)

	return b.PerformPackages(m, func(pkg bootstrapper.Package) error {
		dest, ok := pkgDirs[pkg.Name]

		if !ok {
			return fmt.Errorf("unhandled package: %s", pkg.Name)
		}

		return pkg.Extract(filepath.Join(dirs.Downloads, pkg.Checksum), filepath.Join(b.Dir, dest))
	})
}

func (b *Binary) SetupDxvk() error {
	ver, err := state.DxvkVersion()
	if err != nil {
		return err
	}
	installed := ver != ""

	if installed && !b.GlobalConfig.Player.Dxvk && !b.GlobalConfig.Studio.Dxvk {
		b.Splash.Message("Uninstalling DXVK")
		if err := dxvk.Remove(b.Prefix); err != nil {
			return err
		}

		return state.SaveDxvk("")
	}

	if !b.Config.Dxvk {
		return nil
	}

	b.Splash.Progress(0.0)
	dxvk.Setenv()

	if b.GlobalConfig.DxvkVersion == ver {
		return nil
	}

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return err
	}
	path := filepath.Join(dirs.Cache, "dxvk-"+b.GlobalConfig.DxvkVersion+".tar.gz")

	b.Splash.Progress(0.3)
	b.Splash.Message("Downloading DXVK")
	if err := dxvk.Fetch(path, b.GlobalConfig.DxvkVersion); err != nil {
		return err
	}

	b.Splash.Progress(0.7)
	b.Splash.Message("Extracting DXVK")
	if err := dxvk.Extract(path, b.Prefix); err != nil {
		return err
	}
	b.Splash.Progress(1.0)

	return state.SaveDxvk(b.GlobalConfig.DxvkVersion)
}

func (b *Binary) Command(args ...string) (*wine.Cmd, error) {
	if strings.HasPrefix(strings.Join(args, " "), "roblox-studio:1") {
		args = []string{"-protocolString", args[0]}
	}

	if b.GlobalConfig.MultipleInstances {
		mutexer := b.Prefix.Command("wine", filepath.Join(BinPrefix, "robloxmutexer.exe"))
		err := mutexer.Start()
		if err != nil {
			return &wine.Cmd{}, err
		}
	}

	cmd := b.Prefix.Wine(filepath.Join(b.Dir, b.Type.Executable()), args...)

	launcher := strings.Fields(b.Config.Launcher)
	if len(launcher) >= 1 {
		cmd.Args = append(launcher, cmd.Args...)

		launcherPath, err := exec.LookPath(launcher[0])
		if err != nil {
			return &wine.Cmd{}, err
		}

		cmd.Path = launcherPath
	}

	return cmd, nil
}
