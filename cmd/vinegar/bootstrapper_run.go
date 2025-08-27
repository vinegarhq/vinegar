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

	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/logging"
)

func run(cmd *wine.Cmd) error {
	cmd.Stderr = nil
	out, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		s := bufio.NewScanner(out)
		for s.Scan() {
			line := s.Text()
			slog.Log(context.Background(), logging.LevelWine, line)
		}
	}()

	return cmd.Run()
}

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

	cmd.Stderr = nil
	out, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
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

	if err := cmd.Start(); err != nil {
		return err
	}

	idle(func() {
		b.win.Destroy()
		b.app.Hold()
	})
	defer idle(b.app.Release)

	if err := b.registerGameMode(cmd.Process.Pid); err != nil {
		slog.Error("Failed to register with GameMode", "err", err)
	}

	idle(func() { b.ActivateAction("show-stop", nil) })

	go b.handleWineOutput(out)

	err = cmd.Wait()

	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == -1 {
		signal := cmd.ProcessState.Sys().(syscall.WaitStatus).Signal()
		slog.Warn("Roblox was killed!", "signal", signal)
		return nil
	}

	return err
}

func (b *bootstrapper) handleWineOutput(wr io.Reader) {
	s := bufio.NewScanner(wr)

	for s.Scan() {
		line := s.Text()

		// XXXX:channel:class OutputDebugStringA "[FLog::Foo] Message"
		if len(line) >= 39 && line[19:37] == "OutputDebugStringA" {
			// Avoid "\n" calls to OutputDebugStringA
			if len(line) >= 87 {
				b.handleRobloxLog(line[39 : len(line)-1])
			}
			continue
		}

		slog.Log(context.Background(), logging.LevelWine, line)
	}
}
