package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	bsrpc "github.com/vinegarhq/vinegar/bloxstraprpc"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/internal/bus"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/state"
	"github.com/vinegarhq/vinegar/roblox"
	boot "github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/splash"
	"github.com/vinegarhq/vinegar/util"
	"github.com/vinegarhq/vinegar/wine"
	"github.com/vinegarhq/vinegar/wine/dxvk"
)

const (
	DialogInternalBrowserBrokenTitle = "WebView/InternalBrowser is broken"
	DialogUseBrowserMsg              = "Use the browser for whatever you were doing just now."
	DialogQuickLoginMsg              = "Use Quick Log In to authenticate ('Log In With Another Device' button)"
	DialogNoAVXTitle                 = "Minimum requirements aren't met"
	DialogNoAVXMsg                   = "Your machine's CPU doesn't have AVX extensions, which is a requirement for running Roblox on Linux."
)

type Binary struct {
	Splash *splash.Splash
	State  *state.State

	GlobalConfig *config.Config
	Config       *config.Binary

	Alias  string
	Name   string
	Dir    string
	Prefix *wine.Prefix
	Type   roblox.BinaryType
	Deploy *boot.Deployment

	// Logging
	Auth     bool
	Activity bsrpc.Activity

	// DBUS session
	BusSession *bus.SessionBus
}

func NewBinary(bt roblox.BinaryType, cfg *config.Config, pfx *wine.Prefix) *Binary {
	var bcfg config.Binary

	switch bt {
	case roblox.Player:
		bcfg = cfg.Player
	case roblox.Studio:
		bcfg = cfg.Studio
	}

	return &Binary{
		Activity: bsrpc.New(),
		Splash:   splash.New(&cfg.Splash),

		GlobalConfig: cfg,
		Config:       &bcfg,

		Alias:  bt.String(),
		Name:   bt.BinaryName(),
		Type:   bt,
		Prefix: pfx,

		BusSession: bus.New(),
	}
}

func (b *Binary) Run(args ...string) error {
	if b.Config.DiscordRPC {
		if err := b.Activity.Connect(); err != nil {
			log.Printf("WARNING: Could not initialize Discord RPC: %s, disabling...", err)
			b.Config.DiscordRPC = false
		} else {
			defer b.Activity.Close()
		}
	}

	// REQUIRED for HandleRobloxLog to function.
	os.Setenv("WINEDEBUG", os.Getenv("WINEDEBUG")+",warn+debugstr")

	cmd, err := b.Command(args...)
	if err != nil {
		return err
	}
	o, err := cmd.OutputPipe()
	if err != nil {
		return err
	}

	// Act as the signal holder, as roblox/wine will not do anything with the INT signal.
	// Additionally, if Vinegar got TERM, it will also immediately exit, but roblox
	// continues running.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		// Only kill the process if it even had a PID
		if cmd.Process != nil {
			log.Println("Killing Roblox")
			// This way, cmd.Run() will return and the wineprefix killer will be ran.
			cmd.Process.Kill()
		}

		// Don't handle INT after it was recieved, this way if another signal was sent,
		// Vinegar will immediately exit.
		signal.Stop(c)
	}()

	go b.HandleOutput(o)

	log.Printf("Launching %s (%s)", b.Name, cmd)
	b.Splash.SetMessage("Launching " + b.Alias)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("roblox process: %w", err)
	}

	if b.Config.GameMode {
		if err := b.BusSession.GamemodeRegister(int32(cmd.Process.Pid)); err != nil {
			fmt.Fprintf(os.Stderr, "failed to register gamemode: %s", err.Error())
		}
	}

	defer func() {
		// Don't do anything if the process even ran correctly.
		if cmd.Process == nil {
			return
		}

		for {
			time.Sleep(100 * time.Millisecond)

			// This is because there may be a race condition between the process
			// procfs depletion and the proccess getting killed.
			// CommFound walks over procfs, so here ensure that the process no longer
			// exists in procfs.
			_, err := os.Stat(filepath.Join("/proc", strconv.Itoa(cmd.Process.Pid)))
			if err != nil {
				break
			}
		}

		if util.CommFound("Roblox") {
			log.Println("Another Roblox instance is already running, not killing wineprefix")
			return
		}

		b.Prefix.Kill()
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("roblox process: %w", err)
	}

	return nil
}

func (b *Binary) HandleOutput(wr io.Reader) {
	s := bufio.NewScanner(wr)
	for s.Scan() {
		txt := s.Text()

		// XXXX:channel:class OutputDebugStringA "[FLog::Foo] Message"
		if len(txt) >= 39 && txt[19:37] == "OutputDebugStringA" {
			// length of roblox Flog message
			if len(txt) >= 90 {
				b.HandleRobloxLog(txt[39 : len(txt)-1])
			}
			continue
		}

		fmt.Fprintln(b.Prefix.Output, txt)
	}
}

func (b *Binary) HandleRobloxLog(line string) {
	// As soon as a singular Roblox log has been hit, close the splash window
	if !b.Splash.IsClosed() {
		b.Splash.Close()
	}

	fmt.Fprintln(b.Prefix.Output, line)

	if strings.Contains(line, "DID_LOG_IN") {
		b.Auth = true
		return
	}

	if strings.Contains(line, "InternalBrowser") {
		msg := DialogUseBrowserMsg
		if !b.Auth {
			msg = DialogQuickLoginMsg
		}

		b.Splash.Dialog(DialogInternalBrowserBrokenTitle, msg)
		return
	}

	if b.Config.DiscordRPC {
		if err := b.Activity.HandleRobloxLog(line); err != nil {
			log.Printf("Failed to handle Discord RPC: %s", err)
		}
	}
}

func (b *Binary) FetchDeployment() error {
	b.Splash.SetMessage("Fetching " + b.Alias)

	if b.Config.ForcedVersion != "" {
		log.Printf("WARNING: using forced version: %s", b.Config.ForcedVersion)

		d := boot.NewDeployment(b.Type, b.Config.Channel, b.Config.ForcedVersion)
		b.Deploy = &d
		return nil
	}

	d, err := boot.FetchDeployment(b.Type, b.Config.Channel)
	if err != nil {
		return err
	}

	b.Deploy = &d
	return nil
}

func (b *Binary) Setup() error {
	s, err := state.Load()
	if err != nil {
		return err
	}
	b.State = &s

	if err := b.FetchDeployment(); err != nil {
		return err
	}

	b.Dir = filepath.Join(dirs.Versions, b.Deploy.GUID)
	b.Splash.SetDesc(fmt.Sprintf("%s %s", b.Deploy.GUID, b.Deploy.Channel))

	stateVer := b.State.Version(b.Type)
	if stateVer != b.Deploy.GUID {
		log.Printf("Installing %s (%s -> %s)", b.Name, stateVer, b.Deploy.GUID)

		if err := b.Install(); err != nil {
			return err
		}
	} else {
		log.Printf("%s is up to date (%s)", b.Name, b.Deploy.GUID)
	}

	b.Config.Env.Setenv()

	if err := b.Config.FFlags.Apply(b.Dir); err != nil {
		return err
	}

	if err := dirs.OverlayDir(b.Dir); err != nil {
		return err
	}

	if err := b.SetupDxvk(); err != nil {
		return err
	}

	b.Splash.SetProgress(1.0)
	return b.State.Save()
}

func (b *Binary) Install() error {
	b.Splash.SetMessage("Installing " + b.Alias)

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	manifest, err := boot.FetchPackageManifest(b.Deploy)
	if err != nil {
		return err
	}

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	b.Splash.SetMessage("Downloading " + b.Alias)
	if err := b.DownloadPackages(&manifest); err != nil {
		return err
	}

	b.Splash.SetMessage("Extracting " + b.Alias)
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

	if err := boot.WriteAppSettings(b.Dir); err != nil {
		return err
	}

	b.State.AddBinary(&manifest)

	if err := b.State.CleanPackages(); err != nil {
		return err
	}

	return b.State.CleanVersions()
}

func (b *Binary) PerformPackages(pm *boot.PackageManifest, fn func(boot.Package) error) error {
	donePkgs := 0
	pkgsLen := len(pm.Packages)

	return pm.Packages.Perform(func(pkg boot.Package) error {
		err := fn(pkg)
		if err != nil {
			return err
		}

		donePkgs++
		b.Splash.SetProgress(float32(donePkgs) / float32(pkgsLen))

		return nil
	})
}

func (b *Binary) DownloadPackages(pm *boot.PackageManifest) error {
	log.Printf("Downloading %d Packages for %s", len(pm.Packages), pm.Deployment.GUID)

	return b.PerformPackages(pm, func(pkg boot.Package) error {
		return pkg.Fetch(filepath.Join(dirs.Downloads, pkg.Checksum), pm.DeployURL)
	})
}

func (b *Binary) ExtractPackages(pm *boot.PackageManifest) error {
	log.Printf("Extracting %d Packages for %s", len(pm.Packages), pm.Deployment.GUID)

	pkgDirs := boot.BinaryDirectories(b.Type)

	return b.PerformPackages(pm, func(pkg boot.Package) error {
		dest, ok := pkgDirs[pkg.Name]

		if !ok {
			return fmt.Errorf("unhandled package: %s", pkg.Name)
		}

		return pkg.Extract(filepath.Join(dirs.Downloads, pkg.Checksum), filepath.Join(b.Dir, dest))
	})
}

func (b *Binary) SetupDxvk() error {
	if b.State.DxvkVersion != "" && !b.GlobalConfig.Player.Dxvk && !b.GlobalConfig.Studio.Dxvk {
		if err := dxvk.Remove(b.Prefix); err != nil {
			return err
		}

		b.State.DxvkVersion = ""
		return nil
	}

	if !b.Config.Dxvk {
		return nil
	}

	dxvk.Setenv(b.Prefix)

	if b.GlobalConfig.DxvkVersion == b.State.DxvkVersion {
		return nil
	}

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return err
	}

	b.Splash.SetMessage("Installing DXVK")
	// This would only get saved if Install succeeded
	b.State.DxvkVersion = b.GlobalConfig.DxvkVersion

	return dxvk.Install(b.GlobalConfig.DxvkVersion, b.Prefix)
}

func (b *Binary) Command(args ...string) (*wine.Cmd, error) {
	if strings.HasPrefix(strings.Join(args, " "), "roblox-studio:1") {
		args = []string{"-protocolString", args[0]}
	}

	if b.GlobalConfig.MultipleInstances {
		log.Println("Launching robloxmutexer in background")

		mutexer := b.Prefix.Command("wine", filepath.Join(BinPrefix, "robloxmutexer.exe"))
		err := mutexer.Start()
		if err != nil {
			return &wine.Cmd{}, fmt.Errorf("robloxmutexer: %w", err)
		}
	}

	cmd := b.Prefix.Wine(filepath.Join(b.Dir, b.Type.Executable()), args...)

	launcher := strings.Fields(b.Config.Launcher)
	if len(launcher) >= 1 {
		cmd.Args = append(launcher, cmd.Args...)

		// For safety, ensure that the launcher is in PATH
		launcherPath, err := exec.LookPath(launcher[0])
		if err != nil {
			return &wine.Cmd{}, err
		}

		cmd.Path = launcherPath
	}

	return cmd, nil
}
