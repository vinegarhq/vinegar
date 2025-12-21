package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/gutil"
)

func (b *bootstrapper) command(args ...string) (*wine.Cmd, error) {
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
			slog.Warn("Killing Roblox", "pid", cmd.Process.Pid)
			cmd.Process.Kill()
		}
	}()

	if err := cmd.Start(); err != nil {
		return err
	}
	b.procs = append(b.procs, cmd.Process)
	defer func() {
		b.procs = slices.DeleteFunc(b.procs, func(p *os.Process) bool {
			return p == cmd.Process
		})
		if len(b.procs) > 0 {
			return
		}

		// Workaround any other stray processes holding Wine up
		// such as WebView
		slog.Warn("No more processes left, killing Wineprefix")
		b.pfx.Kill()
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
