package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/apprehensions/wine"
	"github.com/apprehensions/wine/dxvk"
	"github.com/fsnotify/fsnotify"
	"github.com/nxadm/tail"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
)

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

func (b *bootstrapper) Execute(args ...string) error {
	defer Background(b.win.Destroy)

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

func (b *bootstrapper) SetupPrefix() error {
	b.status.SetLabel("Setting up Wine")

	if c := b.pfx.Wine(""); c.Err != nil {
		return fmt.Errorf("wine: %w", c.Err)
	}

	setup := false
	if _, err := os.Stat(filepath.Join(b.pfx.Dir(), "drive_c", "windows")); err != nil {
		setup = true
	}

	if setup {
		return b.PrefixInit()
	}

	return nil
}

func (b *bootstrapper) PrefixInit() error {
	slog.Info("Initializing wineprefix", "dir", b.pfx.Dir())
	b.status.SetLabel("Initializing")

	if err := b.pfx.Init(); err != nil {
		return fmt.Errorf("prefix init: %w", err)
	}

	if err := b.pfx.SetDPI(97); err != nil {
		return fmt.Errorf("prefix set dpi: %w", err)
	}

	if err := b.InstallWebView(); err != nil {
		return fmt.Errorf("prefix webview install: %w", err)
	}

	return nil
}

func (ui *ui) DeletePrefixes() error {
	slog.Info("Deleting Wineprefixes!")

	if err := os.RemoveAll(dirs.Prefixes); err != nil {
		return fmt.Errorf("remove prefixes: %w", err)
	}

	ui.state.Studio.DxvkVersion = ""

	if err := ui.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
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

func (b *bootstrapper) SetupDxvk() error {
	dxvk.Setenv(b.cfg.Studio.Dxvk)

	if !b.cfg.Studio.Dxvk {
		return nil
	}

	b.status.SetLabel("Setting up DXVK")

	if b.cfg.Studio.DxvkVersion == b.state.Studio.DxvkVersion {
		slog.Info("DXVK up to date!", "version", b.state.Studio.DxvkVersion)
		return nil
	}

	if err := b.DxvkInstall(); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	return nil
}

func (b *bootstrapper) DxvkInstall() error {
	ver := b.cfg.Studio.DxvkVersion
	dxvkPath := filepath.Join(dirs.Cache, "dxvk-"+ver+".tar.gz")

	if _, err := os.Stat(dxvkPath); err != nil {
		url := dxvk.URL(ver)

		slog.Info("Downloading DXVK", "ver", ver)
		b.status.SetLabel("Downloading DXVK")

		if err := netutil.DownloadProgress(url, dxvkPath, &b.pbar); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	b.status.SetLabel("Installing DXVK")

	if err := dxvk.Extract(b.pfx, dxvkPath); err != nil {
		return err
	}

	b.state.Studio.DxvkVersion = ver
	return nil
}
