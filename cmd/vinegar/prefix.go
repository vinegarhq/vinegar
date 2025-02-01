package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/apprehensions/wine"
	"github.com/apprehensions/wine/dxvk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/netutil"
)

func WineSimpleRun(cmd *wine.Cmd) error {
	cmd.Stderr = nil
	out, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	s := bufio.NewScanner(out)
	for s.Scan() {
		line := s.Text()
		slog.Log(context.Background(), logging.LevelWine, line)
	}

	return cmd.Wait()
}

func (s *ui) KillPrefix() error {
	slog.Info("Killing Wineprefix...")
	return s.pfx.Kill()
}

func (s *ui) RunWinetricks() error {
	slog.Info("Running Winetricks!")
	return WineSimpleRun(s.pfx.Tricks())
}

func (ui *ui) DeletePrefixes() error {
	slog.Info("Deleting Wineprefixes!")

	if err := ui.KillPrefix(); err != nil {
		return fmt.Errorf("kill prefix: %w", err)
	}

	if err := os.RemoveAll(dirs.Prefixes); err != nil {
		return err
	}

	ui.state.Studio.DxvkVersion = ""

	if err := ui.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
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

func (b *bootstrapper) Execute(args ...string) error {
	cmd, err := b.Command(args...)
	if err != nil {
		return err
	}

	cmd.Stderr = nil
	out, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	// Roblox will keep running if it was sent SIGINT; requiring acting as the signal holder.
	// SIGUSR1 is used in Tail() to force kill roblox, used to differenciate between
	// a user-sent signal and a self sent signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(c)
	go func() {
		s := <-c
		signal.Stop(c)

		slog.Warn("Recieved signal", "signal", s)

		// Only kill Roblox if it hasn't exited
		if cmd.ProcessState == nil {
			slog.Warn("Killing Roblox", "pid", cmd.Process.Pid)
			cmd.Process.Kill()
		}
	}()

	slog.Info("Running Studio", "cmd", cmd)
	b.status.SetLabel("Launching Studio")

	if err := cmd.Start(); err != nil {
		return err
	}

	b.win.Hide()

	if b.cfg.Studio.GameMode {
		b.RegisterGameMode(int32(cmd.Process.Pid))
	}

	b.HandleWineOutput(out)

	err = cmd.Wait()

	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == -1 {
		signal := cmd.ProcessState.Sys().(syscall.WaitStatus).Signal()
		slog.Warn("Roblox was killed!", "signal", signal)
		return nil
	}

	return err
}

func (b *bootstrapper) HandleWineOutput(wr io.Reader) {
	s := bufio.NewScanner(wr)
	closed := false

	for s.Scan() {
		line := s.Text()

		// XXXX:channel:class OutputDebugStringA "[FLog::Foo] Message"
		if len(line) >= 39 && line[19:37] == "OutputDebugStringA" {
			if !closed {
				b.win.Hide()
				closed = true
			}

			b.HandleRobloxLog(line)
		}

		slog.Log(context.Background(), logging.LevelWine, line)
	}
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
	b.Message("Initializing Wineprefix")
	defer b.Performing()()

	if err := WineSimpleRun(b.pfx.Init()); err != nil {
		return fmt.Errorf("prefix init: %w", err)
	}

	b.Message("Setting Wineprefix DPI")
	if err := b.pfx.SetDPI(97); err != nil {
		return fmt.Errorf("prefix set dpi: %w", err)
	}

	b.Message("Setting Wineprefix version")
	if err := WineSimpleRun(b.pfx.Wine("winecfg", "/v", "win10")); err != nil {
		return err
	}

	return nil
}

func (b *bootstrapper) SetupDxvk() error {
	dxvk.Setenv(b.cfg.Studio.Dxvk)

	if !b.cfg.Studio.Dxvk {
		return nil
	}

	b.Message("Setting up DXVK")

	if b.cfg.Studio.DxvkVersion == b.state.Studio.DxvkVersion {
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

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return err
	}

	if _, err := os.Stat(dxvkPath); err != nil {
		url := dxvk.URL(ver)

		slog.Info("Downloading DXVK", "ver", ver)
		b.status.SetLabel("Downloading DXVK")

		if err := netutil.DownloadProgress(url, dxvkPath, &b.pbar); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	b.status.SetLabel("Installing DXVK")
	slog.Info("Extracting DXVK", "version", ver)

	if err := dxvk.Extract(b.pfx, dxvkPath); err != nil {
		return err
	}

	b.state.Studio.DxvkVersion = ver

	return nil
}
