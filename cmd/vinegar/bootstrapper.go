package main

import (
	"errors"
	"fmt"
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
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/nxadm/tail"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/studiorpc"
)

var Studio = clientsettings.WindowsStudio64

const KillWait = 3 * time.Second

const (
	// Randomly chosen log entry in cases where Studios process
	// continues to run. Due to a lack of bug reports, it is unknown
	// specifically which entry to use for these types of cases.
	StudioShutdownEntry = "[FLog::LifecycleManager] Exited ApplicationScope"
)

type bootstrapper struct {
	*ui
	builder *gtk.Builder
	win     *adw.Window

	pbar   gtk.ProgressBar
	status gtk.Label

	dir string
	pfx *wine.Prefix
	bin rbxbin.Deployment

	rp *studiorpc.StudioRPC
}

func (s *ui) NewBootstrapper() bootstrapper {
	b := bootstrapper{
		builder: gtk.NewBuilderFromString(resource("bootstrapper.ui"), -1),
		ui:      s,
		rp:      studiorpc.New(),
		pfx: wine.New(
			filepath.Join(dirs.Prefixes, "studio"),
			s.cfg.Studio.WineRoot,
		),
	}

	provider := gtk.NewCssProvider()
	provider.LoadFromData(style, -1)
	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), provider,
		uint(gtk.STYLE_PROVIDER_PRIORITY_APPLICATION))
	provider.Unref()

	var win adw.Window
	b.builder.GetObject("bootstrapper").Cast(&win)
	b.win = &win
	b.win.SetApplication(&s.app.Application)
	s.app.AddWindow(&b.win.Window)

	var logo gtk.Image
	b.builder.GetObject("logo").Cast(&logo)
	setLogoImage(&logo)
	logo.Unref()

	b.builder.GetObject("status").Cast(&b.status)
	b.builder.GetObject("progress").Cast(&b.pbar)
	b.status.Unref()
	b.pbar.Unref()

	b.win.Present()
	b.win.Unref()

	return b
}

func (b *bootstrapper) Start(args ...string) {
	var idlecb glib.SourceFunc
	idlecb = func(uintptr) bool {
		go func() {
			if err := b.Run(args...); err != nil {
				b.presentError(err)
			}
		}()
		return false
	}
	glib.IdleAdd(&idlecb, 0)
}

func (b *bootstrapper) Run(args ...string) error {
	if err := b.Init(); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	if len(args) == 1 && args[0] == "roblox-" {
		b.HandleProtocolURI(args[0])
	}

	if err := b.Setup(); err != nil {
		return fmt.Errorf("failed to setup roblox: %w", err)
	}

	if err := b.Execute(args...); err != nil {
		return fmt.Errorf("failed to run roblox: %w", err)
	}

	return nil
}

func (b *bootstrapper) Init() error {
	firstRun := false
	if _, err := os.Stat(filepath.Join(b.pfx.Dir(), "drive_c", "windows")); err != nil {
		firstRun = true
	}

	if c := b.pfx.Wine(""); c.Err != nil {
		return fmt.Errorf("wine: %w", c.Err)
	}

	if firstRun {
		slog.Info("Initializing wineprefix", "dir", b.pfx.Dir())
		b.status.SetLabel("Initializing wineprefix")

		// Studio accepts all DPIs except the default, which is 96.
		// Technically this is 'initializing wineprefix', as SetDPI calls Wine which
		// automatically create the Wineprefix.
		err := b.pfx.SetDPI(97)
		if err != nil {
			return fmt.Errorf("pfx init: %w", err)
		}

		if err := b.InstallWebView(); err != nil {
			return fmt.Errorf("webview: %w", err)
		}
	}

	return nil
}

func (b *bootstrapper) HandleProtocolURI(mime string) {
	uris := strings.Split(mime, "+")
	for _, uri := range uris {
		kv := strings.Split(uri, ":")

		if len(kv) == 2 && kv[0] == "channel" {
			c := kv[1]
			if c == "" {
				continue
			}

			slog.Warn("Roblox has requested a user channel, changing...", "channel", c)
			b.cfg.Studio.Channel = c
		}
	}
}

func (b *bootstrapper) Execute(args ...string) error {
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
	b.status.SetLabel("Launching Studio")

	go func() {
		// Wait for process to start
		for {
			if cmd.Process != nil {
				break
			}
		}

		// If the log file wasn't found, assume failure
		// and don't perform post-launch roblox functions.
		lf, err := RobloxLogFile(b.pfx)
		if err != nil {
			slog.Error("Failed to find Roblox log file", "error", err.Error())
			return
		}

		b.win.Hide()

		if b.cfg.Studio.GameMode {
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

func (b *bootstrapper) Tail(name string) {
	t, err := tail.TailFile(name, tail.Config{Follow: true})
	if err != nil {
		slog.Error("Could not tail Roblox log file", "error", err)
		return
	}

	for line := range t.Lines {
		fmt.Fprintln(b.pfx.Stderr, line.Text)

		if strings.Contains(line.Text, StudioShutdownEntry) {
			go func() {
				time.Sleep(KillWait)
				syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			}()
		}

		if b.cfg.Studio.DiscordRPC {
			if err := b.rp.Handle(line.Text); err != nil {
				slog.Error("Presence handling failed", "error", err)
			}
		}
	}
}

func (b *bootstrapper) Command(args ...string) (*wine.Cmd, error) {
	if strings.HasPrefix(strings.Join(args, " "), "roblox-studio:1") {
		args = []string{"-protocolString", args[0]}
	}

	cmd := b.pfx.Wine(filepath.Join(b.dir, "RobloxStudioBeta.exe"), args...)
	if cmd.Err != nil {
		return nil, cmd.Err
	}

	launcher := strings.Fields(b.cfg.Studio.Launcher)
	if len(launcher) >= 1 {
		cmd.Args = append(launcher, cmd.Args...)
		p, err := b.cfg.Studio.LauncherPath()
		if err != nil {
			return nil, fmt.Errorf("bad launcher: %w", err)
		}
		cmd.Path = p
	}

	return cmd, nil
}

func (b *bootstrapper) RegisterGameMode(pid int32) {
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

func LogFile() (*os.File, error) {
	if err := dirs.Mkdirs(dirs.Logs); err != nil {
		return nil, err
	}

	// name-2006-01-02T15:04:05Z07:00.log
	path := filepath.Join(dirs.Logs, time.Now().Format(time.RFC3339)+".log")

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return file, nil
}
