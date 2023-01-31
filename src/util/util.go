// Copyright vinegar-development 2023

package util

import (
	"github.com/adrg/xdg"
	"github.com/hashicorp/go-getter"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Constants
const RBXPLAYERLAUNCHERURL = "https://www.roblox.com/download/client"
const RBXFPSUNLOCKERURL = "https://github.com/axstin/rbxfpsunlocker/releases/download/v4.4.4/rbxfpsunlocker-x64.zip"
const RBXFPSUNLOCKERHASH = "sha256:050fe7c0127dbd4fdc0cecf3ba46248ba7e14d37edba1a54eac40602c130f2f8" // This is going to be a pain to keep updated...

// Variables
type Dirs struct {
	Cache string
	Data  string
	Log   string
	Pfx   string
	Exe   string
}

// Functions
func Errc(e error) { // Thanks, wael.
	if e != nil {
		log.Fatal(e)
	}
}

// Initialize directories
func InitDirs() *Dirs {
	dirs := new(Dirs)
	dirs.Cache = (filepath.Join(xdg.CacheHome, "/vinegar"))
	dirs.Data = (filepath.Join(xdg.DataHome, "/vinegar"))
	dirs.Log = (filepath.Join(dirs.Cache, "/logs"))
	dirs.Pfx = (filepath.Join(dirs.Data, "/pfx"))
	dirs.Exe = (filepath.Join(dirs.Cache, "/exe"))
	Errc(os.MkdirAll(dirs.Cache, 0755))
	Errc(os.MkdirAll(dirs.Data, 0755))
	Errc(os.MkdirAll(dirs.Log, 0755))
	Errc(os.MkdirAll(dirs.Pfx, 0755))
	Errc(os.MkdirAll(dirs.Exe, 0755))
	/*
		Technically this can be done with a loop and the reflect library,
		but it's not worthwhile to replace 5 lines.
	*/
	log.Println("Found necessary directories.")
	return dirs
}

// Setup Environment for wine. Dirs must be inited first.
func InitEnvironment(dirs *Dirs) { // Thanks, wael.
	os.Setenv("WINEPREFIX", dirs.Pfx)
	os.Setenv("WINEARCH", "win64")
	os.Setenv("WINEDEBUG", "fixme-all,-wininet,-ntlm,-winediag,-kerberos")
	os.Setenv("WINEDLLOVERRIDES", "dxdiagn=d;winemenubuilder.exe=d;d3d10core=n;d3d11=n;d3d9=n;dxgi=n")
	os.Setenv("DXVK_LOG_LEVEL", "warn")
	os.Setenv("DXVK_LOG_PATH", "none")
	os.Setenv("DXVK_STATE_CACHE_PATH", filepath.Join(dirs.Cache, "/dxvk"))
}

// Look for FPS unlocker and launcher, and update them as necessary.
// TODO: Figure out a way to check for FPS unlocker updates.
func CheckExecutables(dirs *Dirs) {
	wg := new(sync.WaitGroup)
	// Check if Launcher exists
	if _, err := os.Stat(filepath.Join(dirs.Exe, "RobloxPlayerLauncher.exe")); err == nil {
		log.Println("Found launcher!")
	} else {
		log.Println("No launcher found, installing...")
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := getter.GetFile(filepath.Join(dirs.Exe, "/RobloxPlayerLauncher.exe"), RBXPLAYERLAUNCHERURL); err != nil {
				Errc(err)
			}
		}()
	}

	// TODO: Check for studio here

	// Check if Unlocker exists
	if _, err := os.Stat(filepath.Join(dirs.Exe, "rbxfpsunlocker.exe")); err == nil {
		log.Println("Found FPS unlocker!")
	} else {
		log.Println("No FPS unlocker found, installing...")
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := getter.GetFile(filepath.Join(dirs.Exe, "/rbxfpsunlocker.exe"), (RBXFPSUNLOCKERURL + "?checksum=" + RBXFPSUNLOCKERHASH)); err != nil {
				Errc(err)
				//While technically Roblox can play without rbxfpsunlocker, we may have dependencies on it further in the code
				//Additionally, performance sucks without it.
			}
		}()
	}
	wg.Wait()
}
