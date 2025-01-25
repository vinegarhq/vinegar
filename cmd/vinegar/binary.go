package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/apprehensions/rbxbin"
	"github.com/apprehensions/rbxweb/clientsettings"
	"github.com/apprehensions/wine"
	"github.com/fsnotify/fsnotify"
	"github.com/godbus/dbus/v5"
	"github.com/lmittmann/tint"
	"github.com/nxadm/tail"
	slogmulti "github.com/samber/slog-multi"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/state"
	"github.com/vinegarhq/vinegar/splash"
	"github.com/vinegarhq/vinegar/studiorpc"
	"golang.org/x/term"
)

var Studio = clientsettings.WindowsStudio64

const (
	KillWait = 3 * time.Second
)

const (
	// Randomly chosen log entry in cases where Studios process
	// continues to run. Due to a lack of bug reports, it is unknown
	// specifically which entry to use for these types of cases.
	StudioShutdownEntry = "[FLog::LifecycleManager] Exited ApplicationScope"
)

const (
	DialogUseBrowser = "WebView/InternalBrowser is broken, please use the browser for the action that you were doing."
	DialogQuickLogin = "WebView/InternalBrowser is broken, use Quick Log In to authenticate ('Log In With Another Device' button)"
	DialogFailure    = "Vinegar experienced an error:\n%s"
)

type Binary struct {
	// Only initialized in Main
	Splash *splash.Splash

	State  *state.State
	Config *config.Config

	Dir    string
	Prefix *wine.Prefix
	Deploy rbxbin.Deployment

	// Logging
	Auth     bool
	Presence *studiorpc.StudioRPC
}

func BinaryPrefixDir(bt clientsettings.BinaryType) string {
	return filepath.Join(dirs.Prefixes, strings.ToLower(bt.Short()))
}

func NewBinary(cfg *config.Config) (*Binary, error) {
	s, err := state.Load()
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	rp := studiorpc.New()
	path := filepath.Join(dirs.Prefixes, "studio")
	pfx := wine.New(path, cfg.Studio.WineRoot)

	return &Binary{
		Presence: rp,

		State:  &s,
		Config: cfg,

		Prefix: pfx,
	}, nil
}

func (b *Binary) Main(args ...string) int {
	logFile, err := LogFile("Studio")
	if err != nil {
		slog.Error(fmt.Sprintf("create log file: %s", err))
		return 1
	}
	defer logFile.Close()

	slog.SetDefault(slog.New(slogmulti.Fanout(
		tint.NewHandler(os.Stderr, nil),
		tint.NewHandler(logFile, &tint.Options{NoColor: true}),
	)))

	b.Splash = splash.New(&b.Config.Splash)
	b.Prefix.Stderr = io.MultiWriter(os.Stderr, logFile)
	b.Config.Env.Setenv()

	go func() {
		err := b.Splash.Run()
		if errors.Is(splash.ErrClosed, err) {
			slog.Warn("Splash window closed!")

			// Will tell Run() to immediately kill Roblox, as it handles INT/TERM.
			// Otherwise, it will just with the same appropiate signal.
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return
		}

		// The splash window didn't close cleanly (ErrClosed), an
		// internal error occured, and vinegar cannot continue.
		if err != nil {
			slog.Error(fmt.Sprintf("splash: %s", err))
			logFile.Close()
			os.Exit(1)
		}
	}()

	err = b.Run(args...)
	if err != nil {
		slog.Error(err.Error())

		if b.Config.Splash.Enabled && !term.IsTerminal(int(os.Stderr.Fd())) {
			b.Splash.LogPath = logFile.Name()
			b.Splash.SetMessage("Oops!")
			b.Splash.Dialog(fmt.Sprintf(DialogFailure, err), false, "") // blocks
		}

		return 1
	}

	return 0
}

func (b *Binary) Run(args ...string) error {
	if err := b.Init(); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	if len(args) == 1 && args[0] == "roblox-" {
		b.HandleProtocolURI(args[0])
	}

	b.Splash.SetDesc(b.Config.Studio.Channel)

	if err := b.Setup(); err != nil {
		return fmt.Errorf("failed to setup roblox: %w", err)
	}

	if err := b.Execute(args...); err != nil {
		return fmt.Errorf("failed to run roblox: %w", err)
	}

	return nil
}

func (b *Binary) Init() error {
	firstRun := false
	if _, err := os.Stat(filepath.Join(b.Prefix.Dir(), "drive_c", "windows")); err != nil {
		firstRun = true
	}

	if c := b.Prefix.Wine(""); c.Err != nil {
		return fmt.Errorf("wine: %w", c.Err)
	}

	// Command-line flag vs wineprefix initialized
	if firstRun || FirstRun {
		slog.Info("Initializing wineprefix", "dir", b.Prefix.Dir())
		b.Splash.SetMessage("Initializing wineprefix")

		// Studio accepts all DPIs except the default, which is 96.
		// Technically this is 'initializing wineprefix', as SetDPI calls Wine which
		// automatically create the Wineprefix.
		err := b.Prefix.SetDPI(97)
		if err != nil {
			return fmt.Errorf("pfx init: %w", err)
		}

		if err := b.InstallWebView(); err != nil {
			return fmt.Errorf("webview: %w", err)
		}
	}

	return nil
}

func (b *Binary) HandleProtocolURI(mime string) {
	uris := strings.Split(mime, "+")
	for _, uri := range uris {
		kv := strings.Split(uri, ":")

		if len(kv) == 2 && kv[0] == "channel" {
			c := kv[1]
			if c == "" {
				continue
			}

			slog.Warn("Roblox has requested a user channel, changing...", "channel", c)
			b.Config.Studio.Channel = c
		}
	}
}

func (b *Binary) Execute(args ...string) error {
	cmd, err := b.Command(args...)
	if err != nil {
		return err
	}

	// Roblox will keep running if it was sent SIGINT; requiring acting as the signal holder.
	// SIGUSR1 is used in Tail() to force kill roblox, used to differenciate between
	// a user-sent signal and a self sent signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c

		slog.Warn("Recieved signal", "signal", s)

		// Only kill Roblox if it hasn't exited
		if cmd.ProcessState == nil {
			slog.Warn("Killing Roblox", "pid", cmd.Process.Pid)
			// This way, cmd.Run() will return and vinegar (should) exit.
			cmd.Process.Kill()
		}

		// Don't handle INT after it was recieved, this way if another signal was sent,
		// Vinegar will immediately exit.
		signal.Stop(c)
	}()

	slog.Info("Running Studio", "cmd", cmd)
	b.Splash.SetMessage("Launching Studio")

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
			slog.Error("Failed to find Roblox log file", "error", err.Error())
			return
		}

		b.Splash.Close()

		if b.Config.Studio.GameMode {
			b.RegisterGameMode(int32(cmd.Process.Pid))
		}

		// Blocks and tails file forever until roblox is dead, unless
		// if finding the log file had failed.
		b.Tail(lf)
	}()

	err = cmd.Run()
	// Thanks for your time, fizzie on #go-nuts.
	// Signal errors are not handled as errors since they are
	// used internally to kill Roblox as well.
	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == -1 {
		signal := cmd.ProcessState.Sys().(syscall.WaitStatus).Signal()
		slog.Warn("Roblox was killed!", "signal", signal)
		return nil
	}
	return err
}

func RobloxLogFile(pfx *wine.Prefix) (string, error) {
	ad, err := pfx.AppDataDir()
	if err != nil {
		return "", fmt.Errorf("get appdata: %w", err)
	}

	dir := filepath.Join(ad, "Local", "Roblox", "logs")

	// This is required due to fsnotify requiring the directory
	// to watch to exist before adding it.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create roblox log dir: %w", err)
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return "", fmt.Errorf("make fsnotify watcher: %w", err)
	}
	defer w.Close()

	if err := w.Add(dir); err != nil {
		return "", fmt.Errorf("watch roblox log dir: %w", err)
	}

	for {
		select {
		case e := <-w.Events:
			if e.Has(fsnotify.Create) {
				return e.Name, nil
			}
		case err := <-w.Errors:
			slog.Error("Recieved fsnotify watcher error", "error", err)
		}
	}
}

func (b *Binary) Tail(name string) {
	t, err := tail.TailFile(name, tail.Config{Follow: true})
	if err != nil {
		slog.Error("Could not tail Roblox log file", "error", err)
		return
	}

	for line := range t.Lines {
		fmt.Fprintln(b.Prefix.Stderr, line.Text)

		if strings.Contains(line.Text, StudioShutdownEntry) {
			go func() {
				time.Sleep(KillWait)
				syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			}()
		}

		if b.Config.Studio.DiscordRPC {
			if err := b.Presence.Handle(line.Text); err != nil {
				slog.Error("Presence handling failed", "error", err)
			}
		}
	}
}

func (b *Binary) Command(args ...string) (*wine.Cmd, error) {
	if strings.HasPrefix(strings.Join(args, " "), "roblox-studio:1") {
		args = []string{"-protocolString", args[0]}
	}

	cmd := b.Prefix.Wine(filepath.Join(b.Dir, "RobloxStudioBeta.exe"), args...)
	if cmd.Err != nil {
		return nil, cmd.Err
	}

	launcher := strings.Fields(b.Config.Studio.Launcher)
	if len(launcher) >= 1 {
		cmd.Args = append(launcher, cmd.Args...)
		p, err := b.Config.Studio.LauncherPath()
		if err != nil {
			return nil, fmt.Errorf("bad launcher: %w", err)
		}
		cmd.Path = p
	}

	return cmd, nil
}

func (b *Binary) RegisterGameMode(pid int32) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		slog.Error("Failed to connect to D-Bus", "error", err)
		return
	}

	desktop := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")

	call := desktop.Call("org.freedesktop.portal.GameMode.RegisterGame", 0, pid)
	if call.Err != nil && !errors.Is(call.Err, dbus.ErrMsgNoObject) {
		slog.Error("Failed to register to GameMode", "error", err)
		return
	}
}

func LogFile(name string) (*os.File, error) {
	if err := dirs.Mkdirs(dirs.Logs); err != nil {
		return nil, err
	}

	// name-2006-01-02T15:04:05Z07:00.log
	path := filepath.Join(dirs.Logs, name+"-"+time.Now().Format(time.RFC3339)+".log")

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s log file: %w", name, err)
	}

	slog.Info("Logging to file", "path", path)

	return file, nil
}
