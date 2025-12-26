package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v80/github"
	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/netutil"
)

func (a *app) updateWine() error {
	client := github.NewClient(nil)
	ctx := context.Background()

	release, _, err := client.Repositories.GetLatestRelease(ctx, "vinegarhq", "kombucha")
	if err != nil {
		return fmt.Errorf("release: %w", err)
	}

	tag := release.GetTagName()
	dir := filepath.Join(dirs.Data, "kombucha-"+tag)

	slog.Info("Fetched Wine build",
		"tag", tag, "released", release.PublishedAt.Time)

	if _, err := os.Stat(dirs.WinePath); err == nil {
		path, err := os.Readlink(dirs.WinePath)
		if err != nil {
			return fmt.Errorf("readlink: %w", err)
		}
		if filepath.Base(path) == filepath.Base(dir) {
			slog.Info("Wine build up to date", "link", path)
			return nil
		}

		slog.Info("Removing old Wine build", "name", path)
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove build: %w", err)
		}

		if err := os.Remove(dirs.WinePath); err != nil {
			return fmt.Errorf("remove link: %w", err)
		}
	}

	if len(release.Assets) == 0 &&
		release.Assets[0].GetContentType() == "application/x-xz" {
		return errors.New("expected .tar.xz release")
	}
	url := release.Assets[0].GetBrowserDownloadURL()

	slog.Info("Downloading Wine build", "url", url)
	if err := netutil.ExtractURL(url, dirs.Data); err != nil {
		return err
	}

	if err := os.Symlink("kombucha-"+tag, dirs.WinePath); err != nil {
		return fmt.Errorf("create link: %w", err)
	}

	slog.Info("Updated local Wine installation", "tag", tag)
	return nil
}

func (a *app) Write(b []byte) (int, error) {
	for line := range strings.SplitSeq(string(b[:len(b)-1]), "\n") {
		// XXXX:channel:class OutputDebugStringA "[FLog::Foo] Message"
		if a.boot != nil && len(line) >= 39 && line[19:37] == "OutputDebugStringA" {
			// Avoid "\n" calls to OutputDebugStringA
			if len(line) >= 44 {
				a.boot.handleRobloxLog(line[39 : len(line)-1])
			}
			return len(b), nil
		}

		a.handleWineLog(line)
	}
	return len(b), nil
}

func (a *app) handleWineLog(line string) {
	if strings.Contains(line, "to unimplemented function advapi32.dll.SystemFunction036") {
		err := errors.New("Your Wineprefix is corrupt! Please delete all data in Vinegar's settings.")
		gutil.IdleAdd(func() {
			a.pfx.Server(wine.ServerKill, "9")
			a.showError(err)
		})
	}

	slog.Log(context.Background(), logging.LevelWine.Level(), line)
}
