package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/gutil"
)

func (b *bootstrapper) command(args ...string) (*wine.Cmd, error) {
	cmd := b.pfx.Wine(filepath.Join(b.dir, "RobloxStudioBeta.exe"), args...)
	if cmd.Err != nil {
		return nil, cmd.Err
	}

	// This is an authentication call, which is ran to the main Studio instance,
	// no point to run this with the launcher or seperate desktop.
	if len(args) > 0 && strings.HasPrefix(args[0], "roblox-studio-auth:") {
		return cmd, nil
	}

	// I was called a "noob" for my implementation by someone who creates
	// an entirely new Wineprefix for this feature.
	if d := string(b.cfg.Studio.Desktop); d != "" {
		cmd.Args = append([]string{
			"explorer", "/desktop=" + glib.UuidStringRandom() + "," + d,
		}, cmd.Args[1:]...)
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

func (b *bootstrapper) execute(args ...string) error {
	cmd, err := b.command(args...)
	if err != nil {
		return err
	}

	slog.Info("Running Studio!", "cmd", cmd)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(c)
	go func() {
		s := <-c
		signal.Stop(c)

		slog.Warn("Received signal", "signal", s)

		// Only kill Roblox if it hasn't exited
		if cmd.ProcessState == nil {
			slog.Debug("Killing Roblox", "pid", cmd.Process.Pid)
			cmd.Process.Kill()
		}
	}()

	if err := cmd.Start(); err != nil {
		return err
	}

	b.count++
	defer func() {
		b.count--
	}()

	gutil.IdleAdd(func() {
		b.win.SetVisible(false)
	})

	if err := b.registerGameMode(cmd.Process.Pid); err != nil {
		slog.Error("Failed to register with GameMode", "err", err)
	}

	err = cmd.Wait()

	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == -1 {
		signal := cmd.ProcessState.Sys().(syscall.WaitStatus).Signal()
		slog.Warn("Roblox was killed!", "signal", signal)
		return nil
	}

	return err
}
