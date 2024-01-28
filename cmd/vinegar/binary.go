package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nxadm/tail"
	bsrpc "github.com/vinegarhq/vinegar/bloxstraprpc"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/internal/bus"
	"github.com/vinegarhq/vinegar/internal/state"
	"github.com/vinegarhq/vinegar/roblox"
	boot "github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/splash"
	"github.com/vinegarhq/vinegar/sysinfo"
	"github.com/vinegarhq/vinegar/util"
	"github.com/vinegarhq/vinegar/wine"
)

const timeout = 6 * time.Second

const (
	DialogUseBrowser = "WebView/InternalBrowser is broken, please use the browser for the action that you were doing."
	DialogQuickLogin = "WebView/InternalBrowser is broken, use Quick Log In to authenticate ('Log In With Another Device' button)"
	DialogFailure    = "Vinegar experienced an error:\n%s"
	DialogReqChannel = "Roblox is attempting to set your channel to %[1]s, however the current preferred channel is %s.\n\nWould you like to set the channel to %[1]s temporarily?"
	DialogNoWine     = "Wine is required to run Roblox on Linux, please install it appropiate to your distribution."
	DialogNoAVX      = "Warning: Your CPU does not support AVX. While some people may be able to run without it, most are not able to. VinegarHQ cannot provide support for your installation. Continue?"
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

func (b *Binary) Main(args ...string) {
	b.Config.Env.Setenv()

	logFile, err := LogFile(b.Type.String())
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	logOutput := io.MultiWriter(logFile, os.Stderr)
	b.Prefix.Output = logOutput
	log.SetOutput(logOutput)

	firstRun := false
	if _, err := os.Stat(filepath.Join(b.Prefix.Dir(), "drive_c", "windows")); err != nil {
		firstRun = true
	}

	if firstRun && !sysinfo.CPU.AVX {
		c := b.Splash.Dialog(DialogNoAVX, true)
		if !c {
			log.Fatal("avx is (may be) required to run roblox")
		}
		log.Println("WARNING: Running roblox without AVX!")
	}

	if !wine.WineLook() {
		b.Splash.Dialog(DialogNoWine, false)
		log.Fatalf("%s is required to run roblox", wine.Wine)
	}

	go func() {
		err := b.Splash.Run()
		if errors.Is(splash.ErrClosed, err) {
			log.Printf("Splash window closed!")

			// Will tell Run() to immediately kill Roblox, as it handles INT/TERM.
			// Otherwise, it will just with the same appropiate signal.
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return
		}

		// The splash window didn't close cleanly (ErrClosed), an
		// internal error occured.
		if err != nil {
			log.Fatalf("splash: %s", err)
		}
	}()

	errHandler := func(err error) {
		if !b.GlobalConfig.Splash.Enabled || b.Splash.IsClosed() {
			log.Fatal(err)
		}

		log.Println(err)
		b.Splash.LogPath = logFile.Name()
		b.Splash.Invalidate()
		b.Splash.Dialog(fmt.Sprintf(DialogFailure, err), false)
		os.Exit(1)
	}

	// Technically this is 'initializing wineprefix', as SetDPI calls Wine which
	// automatically create the Wineprefix.
	if firstRun {
		log.Printf("Initializing wineprefix at %s", b.Prefix.Dir())
		b.Splash.SetMessage("Initializing wineprefix")

		if err := b.Prefix.SetDPI(97); err != nil {
			b.Splash.SetMessage(err.Error())
			errHandler(err)
		}
	}

	// If the launch uri contains a channel key with a value
	// that isn't empty, Roblox requested a specific channel
	func() {
		if len(args) < 1 {
			return
		}

		c := regexp.MustCompile(`channel:([^+]*)`).FindStringSubmatch(args[0])
		if len(c) < 1 {
			return
		}

		if c[1] != "" && c[1] != b.Config.Channel {
			r := b.Splash.Dialog(
				fmt.Sprintf(DialogReqChannel, c[1], b.Config.Channel),
				true,
			)
			if r {
				log.Println("Switching user channel temporarily to", c[1])
				b.Config.Channel = c[1]
			}
		}
	}()

	b.Splash.SetDesc(b.Config.Channel)

	if err := b.Setup(); err != nil {
		b.Splash.SetMessage("Failed to setup Roblox")
		errHandler(err)
	}

	if err := b.Run(args...); err != nil {
		b.Splash.SetMessage("Failed to run Roblox")
		errHandler(err)
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

	cmd, err := b.Command(args...)
	if err != nil {
		return err
	}

	// Act as the signal holder, as roblox/wine will not do anything with the INT signal.
	// Additionally, if Vinegar got TERM, it will also immediately exit, but roblox
	// continues running if the signal holder was not present.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		// Only kill Roblox if it hasn't exited
		if cmd.ProcessState == nil {
			// This way, cmd.Run() will return and the wineprefix killer will be ran.
			cmd.Process.Kill()
		}

		// Don't handle INT after it was recieved, this way if another signal was sent,
		// Vinegar will immediately exit.
		signal.Stop(c)
	}()

	log.Printf("Launching %s (%s)", b.Name, cmd)
	b.Splash.SetMessage("Launching " + b.Alias)

	go func() {
		// Wait for process to start
		for {
			if cmd.Process != nil {
				break
			}
		}

		// If the log file wasn't found, assume failure
		// and don't perform post-launch roblox functions.
		lf, err := RobloxLogFile(b.Prefix)
		if err != nil {
			log.Println(err)
			return
		}

		b.Splash.Close()

		if b.Config.GameMode {
			if err := b.BusSession.GamemodeRegister(int32(cmd.Process.Pid)); err != nil {
				log.Println("Attempted to register to Gamemode daemon")
			}
		}

		// Blocks and tails file forever until roblox is dead
		if err := b.Tail(lf); err != nil {
			log.Println(err)
		}
	}()

	if err := cmd.Run(); err != nil {
		if strings.Contains(err.Error(), "signal:") {
			log.Println("WARNING: Roblox exited with", err)
			return nil
		}

		return fmt.Errorf("roblox process: %w", err)
	}

	// may or may not prevent a race condition in procfs
	syscall.Sync()

	if util.CommFound("Roblox") {
		log.Println("Another Roblox instance is already running, not killing wineprefix")
		return nil
	}

	b.Prefix.Kill()

	return nil
}

func RobloxLogFile(pfx *wine.Prefix) (string, error) {
	ad, err := pfx.AppDataDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(ad, "Local", "Roblox", "logs")

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return "", err
	}
	defer w.Close()

	if err := w.Add(dir); err != nil {
		return "", err
	}

	t := time.NewTimer(timeout)

	for {
		select {
		case <-t.C:
			return "", fmt.Errorf("roblox log file not found after %s", timeout)
		case e := <-w.Events:
			if e.Has(fsnotify.Create) {
				return e.Name, nil
			}
		case err := <-w.Errors:
			log.Println("fsnotify watcher:", err)
		}
	}
}

func (b *Binary) Tail(name string) error {
	t, err := tail.TailFile(name, tail.Config{Follow: true})
	if err != nil {
		return err
	}

	for line := range t.Lines {
		fmt.Fprintln(b.Prefix.Output, line.Text)

		if b.Config.DiscordRPC {
			if err := b.Activity.HandleRobloxLog(line.Text); err != nil {
				log.Printf("Failed to handle Discord RPC: %s", err)
			}
		}
	}

	return nil
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
		p, err := b.Config.LauncherPath()
		if err != nil {
			return &wine.Cmd{}, err
		}
		cmd.Path = p
	}

	return cmd, nil
}
