package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/google/go-github/v80/github"
	"github.com/sewnie/wine"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/netutil"
)

// Reports whether a Wine Prefix was initialized.
func (a *app) prepareWine() (bool, error) {
	firstRun := !a.pfx.Exists()

	_, err := os.Stat(dirs.WinePath)
	// Check against symlink in case the default is empty (musl)
	if string(a.cfg.Studio.WineRoot) == dirs.WinePath && err != nil {
		if err := a.updateWine(); err != nil {
			return false, fmt.Errorf("dl: %w", err)
		}
	}
	if a.pfx.Running() {
		return false, nil
	}

	a.boot.message("Setting up Wine", "first-time", firstRun)

	if err := a.pfx.Prepare(); err != nil {
		return firstRun, err
	}
	a.updateWineTheme()

	// Do _not_ do this on prefixes that already exist, only new ones,
	// as to not discard the existing appdata directory.
	if !firstRun {
		return false, nil
	}

	folders := wine.NewRegistryKey(
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`)
	folders.SetValue("Local AppData", dirs.Windows(dirs.AppDataPath))
	folders.SetValue("Documents", dirs.Windows(xdg.UserDirs.Documents))
	folders.SetValue("My Pictures", dirs.Windows(xdg.UserDirs.Pictures))

	a.boot.message("Updating Paths")
	err = a.pfx.RegistryImportKey(folders)
	if err != nil {
		return true, fmt.Errorf("paths: %w", err)
	}

	// Required for app data to propagate, and also prepares
	// the application environment.
	if err := a.pfx.Boot(wine.BootRestart).Run(); err != nil {
		return true, fmt.Errorf("restart: %w", err)
	}

	if err := a.boot.restoreSettings(); err != nil {
		return true, fmt.Errorf("restore: %w", err)
	}

	return true, nil
}

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
			a.mgr.showToast("Up to date")
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

	if _, err := os.Stat(dir); err == nil {
		slog.Info("Wine build already present!")
		goto install
	}

	if len(release.Assets) == 0 &&
		release.Assets[0].GetContentType() == "application/x-xz" {
		return errors.New("expected .tar.xz release")
	}

	{
		url := release.Assets[0].GetBrowserDownloadURL()

		slog.Info("Downloading Wine build", "url", url)
		if err := netutil.ExtractURL(url, dirs.Data); err != nil {
			return err
		}
	}

	if a.mgr != nil {
		a.mgr.showToast("Updated Wine")
	}

install:
	if err := os.Symlink("kombucha-"+tag, dirs.WinePath); err != nil {
		return fmt.Errorf("create link: %w", err)
	}

	slog.Info("Set local Wine installation", "tag", tag)

	// re-initializes the wine prefix struct
	a.applyConfig()

	if a.cfg.Studio.WineRoot.IsDefault() {
		return nil
	}
	// BUG: this is would not be reflected in the manager
	a.cfg.Studio.WineRoot.SetDefault()

	if err := a.cfg.Save(); err != nil {
		return fmt.Errorf("update config: %w", err)
	}
	return nil
}

// Implements io.Writer for reading the log from Wine
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

func (a *app) updateWineTheme() {
	// If the studio theme is "Default", the wine theme change will effect
	// studio as well.

	if !a.pfx.Running() {
		slog.Debug("Not changing theme: Wine is not running")
	}
	root := wine.RegistryKey{Name: "HKEY_CURRENT_USER"}
	v := root.Add(`Software\Microsoft\Windows\CurrentVersion`)
	pz := v.Add(`Themes\Personalize`)
	mgr := v.Add(`ThemeManager`)

	mgr.SetValue("ColorName", nil)
	mgr.SetValue("DllName", nil)
	mgr.SetValue("LoadedBefore", nil)
	mgr.SetValue("SizeName", nil)

	// Reloading an MSStyle requires calling windows API, hence why
	// I try to stick to only changing the current theme to none
	// and change the colors. This means that the light theme's clean
	// look will be gone in the light theme.
	mgr.SetValue("LoadedBefore", "0")
	mgr.SetValue("ThemeActive", "0")
	if a.GetStyleManager().GetDark() {
		root.Add(`Control Panel\Colors`).Values = darkThemeValues
		pz.SetValue("AppsUseLightTheme", uint32(0))
		pz.SetValue("SystemUsesLightTheme", uint32(0))
	} else {
		root.Add(`Control Panel\Colors`).Values = lightThemeValues
		pz.SetValue("AppsUseLightTheme", uint32(1))
		pz.SetValue("SystemUsesLightTheme", uint32(1))
	}

	slog.Debug("Changing Wine's theme", "dark", a.GetStyleManager().GetDark())
	if err := a.pfx.RegistryImportKey(&root); err != nil {
		slog.Error("Failed to change Wine's theme", "err", err)
	}

	// Modifying the registry manually triggers no events. To make Studio
	// refresh its theme, run a small application to open a GUI for just a
	// second. Yes this results in a flicker and requires the user to
	// focus on Studio, but there's no other way to allow live updates :/
	if len(a.boot.procs) > 0 {
		_ = a.pfx.Wine("start", "cmd", "/c", "exit").Run()
	}
}

var darkThemeValues = []wine.RegistryValue{
	{"ActiveBorder", "34 34 38"},            // #222226 window_bg_color
	{"ActiveTitle", "46 46 50"},             // #2e2e32 headerbar_bg_color
	{"AppWorkSpace", "45 45 49"},            // midpoint between window/view background
	{"Background", "29 29 32"},              // #1d1d20 view_bg_color
	{"ButtonAlternativeFace", "98 160 234"}, // accent blue from Adwaita
	{"ButtonDkShadow", "22 22 25"},          // slightly darker than view_bg
	{"ButtonFace", "46 46 50"},              // #2e2e32 headerbar
	{"ButtonHilight", "57 57 61"},           // lifted tone from card_bg_color overlay
	{"ButtonLight", "53 53 57"},
	{"ButtonShadow", "25 25 28"},
	{"ButtonText", "255 255 255"},
	{"GradientActiveTitle", "46 46 50"},
	{"GradientInactiveTitle", "40 40 44"}, // sidebar_backdrop_color #28282c
	{"GrayText", "136 136 136"},
	{"Hilight", "98 160 234"}, // accent blue
	{"HilightText", "255 255 255"},
	{"InactiveBorder", "40 40 44"}, // sidebar_backdrop
	{"InactiveTitle", "40 40 44"},
	{"InactiveTitleText", "211 211 211"},
	{"InfoText", "255 255 255"},
	{"InfoWindow", "54 54 58"}, // dialog_bg_color #36363a
	{"Menu", "54 54 58"},       // dialog/popover background
	{"MenuBar", "46 46 50"},    // keep menu slightly lighter
	{"MenuHilight", "98 160 234"},
	{"MenuText", "255 255 255"},
	{"Scrollbar", "46 46 50"},
	{"TitleText", "255 255 255"},
	{"Window", "29 29 32"},      // view_bg_color
	{"WindowFrame", "46 46 50"}, // subtle outline color
	{"WindowText", "255 255 255"},
}

var lightThemeValues = []wine.RegistryValue{
	{"ActiveBorder", "255 255 255"},
	{"ActiveTitle", "50 150 250"},
	{"AppWorkSpace", "128 128 128"},
	{"Background", "37 111 149"},
	{"ButtonAlternateFace", "255 255 255"},
	{"ButtonAlternativeFace", "98 160 234"},
	{"ButtonDkShadow", "106 106 106"},
	{"ButtonFace", "245 245 245"},
	{"ButtonHilight", "255 255 255"},
	{"ButtonLight", "227 227 227"},
	{"ButtonShadow", "166 166 166"},
	{"ButtonText", "0 0 0"},
	{"GradientActiveTitle", "50 150 250"},
	{"GradientInactiveTitle", "128 128 128"},
	{"GrayText", "106 106 106"},
	{"Hilight", "48 150 250"},
	{"HilightText", "255 255 255"},
	{"HotTrackingColor", "48 150 250"},
	{"InactiveBorder", "255 255 255"},
	{"InactiveTitle", "128 128 128"},
	{"InactiveTitleText", "200 200 200"},
	{"InfoText", "0 0 0"},
	{"InfoWindow", "255 255 255"},
	{"Menu", "255 255 255"},
	{"MenuBar", "255 255 255"},
	{"MenuHilight", "48 150 250"},
	{"MenuText", "0 0 0"},
	{"Scrollbar", "255 255 255"},
	{"TitleText", "0 0 0"},
	{"Window", "255 255 255"},
	{"WindowFrame", "158 158 158"},
	{"WindowText", "0 0 0"},
}
