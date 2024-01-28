package main

import (
	"bufio"
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

const (
	DialogUseBrowser = "WebView/InternalBrowser is broken, please use the browser for the action that you were doing."
	DialogQuickLogin = "WebView/InternalBrowser is broken, use Quick Log In to authenticate ('Log In With Another Device' button)"
	DialogFailure    = "Vinegar experienced an error:\n%s"
	DialogReqChannel = "Roblox is attempting to set your channel to %[1]s, however the current preferred channel is %s.\n\nWould you like to set the channel to %[1]s temporarily?"
	DialogNoWine     = "Wine is required to run Roblox on Linux, please install it appropiate to your distribution."
	DialogNoAVX      = "Warning: Your CPU does not support AVX. While some people may be able to run without it, most are not able to. VinegarHQ cannot provide support for your installation. Continue?"
	DialogMerlin     = `VinegarHQ is running an automated survey to better understand users' system details. These include:
• CPU make and model
• GPU make and model
• Kernel version
• Distro name

No personally identifiable information is sent, and the source code of the web server can be found on our GitHub under the "Merlin" repository.

If you would like to help us understand our community better, please participate in our hardware survey! This message won't show again after closing it.

Thank you for using Vinegar.`
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

	go b.HandleOutput(o)

	log.Printf("Launching %s (%s)", b.Name, cmd)
	b.Splash.SetMessage("Launching " + b.Alias)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("roblox process: %w", err)
	}

	if b.Config.GameMode {
		if err := b.BusSession.GamemodeRegister(int32(cmd.Process.Pid)); err != nil {
			log.Println("Attempted to register to Gamemode daemon")
		}
	}

	defer func() {
		// may or may not prevent a race condition in procfs
		syscall.Sync()

		if util.CommFound("Roblox") {
			log.Println("Another Roblox instance is already running, not killing wineprefix")
			return
		}

		b.Prefix.Kill()
	}()

	err = cmd.Wait()
	if err == nil {
		return nil
	}

	// Roblox was sent a signal, do not consider it an error.
	if strings.Contains(err.Error(), "signal:") {
		log.Println("WARNING: Roblox exited with", err)
		return nil
	}

	return fmt.Errorf("roblox: %w", err)
}

func (b *Binary) HandleOutput(wr io.Reader) {
	s := bufio.NewScanner(wr)
	closed := false

	for s.Scan() {
		txt := s.Text()

		// XXXX:channel:class OutputDebugStringA "[FLog::Foo] Message"
		if len(txt) >= 39 && txt[19:37] == "OutputDebugStringA" {
			// As soon as a singular Roblox log has been hit, close the splash window
			if !closed {
				b.Splash.Close()
			}

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
	fmt.Fprintln(b.Prefix.Output, line)

	if strings.Contains(line, "DID_LOG_IN") {
		b.Auth = true
		return
	}

	if strings.Contains(line, "InternalBrowser") {
		msg := DialogUseBrowser
		if !b.Auth {
			msg = DialogQuickLogin
		}

		b.Splash.Dialog(msg, false)
		return
	}

	if b.Config.DiscordRPC {
		if err := b.Activity.HandleRobloxLog(line); err != nil {
			log.Printf("Failed to handle Discord RPC: %s", err)
		}
	}
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
