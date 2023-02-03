// Copyright vinegar-development 2023

package util

import (
	"log"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-getter"
)

const (
	FPSUNLOCKERHASH = "sha256:050fe7c0127dbd4fdc0cecf3ba46248ba7e14d37edba1a54eac40602c130f2f8"
	FPSUNLOCKERLITURL = "https://github.com/axstin/rbxfpsunlocker/releases/download/v4.4.4/rbxfpsunlocker-x64.zip"

	// go-getter functionality
	FPSUNLOCKERURL = FPSUNLOCKERLITURL + "?checksum=" + FPSUNLOCKERHASH
)

type Dirs struct {
	Cache string
	Data  string
	Log   string
	Pfx   string
	Exe   string
}

func Errc(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func DeleteDir(dir ...string) {
	for _, d := range dir {
		log.Println("Deleting directory:", d)
		Errc(os.RemoveAll(d))
	}
}

func DirsCheck(dir ...string) {
	for _, d := range dir {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			log.Println("Creating directory:", d)
		} else { 
			continue
		}
		Errc(os.MkdirAll(d, 0755))
	}
}

func InitDirs() *Dirs {
	dirs := new(Dirs)
	home := os.Getenv("HOME")

	if home == "" {
		log.Fatal("Failed to get $HOME variable")
	}

	dirs.Cache = (filepath.Join(home, ".cache", "/vinegar"))
	dirs.Data  = (filepath.Join(home, ".local", "share", "/vinegar"))
	dirs.Log   = (filepath.Join(dirs.Cache, "/logs"))
	dirs.Pfx   = (filepath.Join(dirs.Data, "/pfx"))
	dirs.Exe   = (filepath.Join(dirs.Cache, "/exe"))

	return dirs
}

func InitEnv(dirs *Dirs) {
	os.Setenv("WINEPREFIX", dirs.Pfx)
	os.Setenv("WINEARCH", "win64")
	// Removal of most unnecessary Wine facilities
	os.Setenv("WINEDEBUG", "fixme-all,-wininet,-ntlm,-winediag,-kerberos")
	os.Setenv("WINEDLLOVERRIDES", "dxdiagn=d;winemenubuilder.exe=d")
	
	os.Setenv("DXVK_LOG_LEVEL", "warn")
	os.Setenv("DXVK_LOG_PATH", "none")
	os.Setenv("DXVK_STATE_CACHE_PATH", filepath.Join(dirs.Cache, "dxvk"))
	
	os.Setenv("MESA_GL_VERSION_OVERRIDE", "4.4")
	// Use the dedicated gpu if available, untested
	os.Setenv("DRI_PRIME", "1")
	os.Setenv("__NV_PRIME_RENDER_OFFLOAD", "1")
	os.Setenv("__VK_LAYER_NV_optimus", "NVIDIA_only")
	os.Setenv("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
}

func Exec(dirs *Dirs, prog string, args ...string) {
	log.Println(args)
	cmd := exec.Command(prog, args...)
	timeFmt := time.Now().Format(time.RFC3339)

	stdoutFile, err := os.Create(filepath.Join(dirs.Log, timeFmt + "-stdout.log"))
	Errc(err)
	log.Println("Forwarding stdout to", stdoutFile.Name())
	defer stdoutFile.Close()

	stderrFile, err := os.Create(filepath.Join(dirs.Log, timeFmt + "-stderr.log"))
	Errc(err)
	log.Println("Forwarding stderr to", stderrFile.Name())
	defer stderrFile.Close()
	
	cmd.Dir = dirs.Cache
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	Errc(cmd.Run())

	logFile, err := stderrFile.Stat()
	Errc(err)
	if logFile.Size() == 0 {
		log.Println("Empty stderr log file detected, deleting")
		Errc(os.RemoveAll(stderrFile.Name()))
	}

	logFile, err = stdoutFile.Stat()
	Errc(err)
	if logFile.Size() == 0 {
		log.Println("Empty stdout file detected, deleted")
		Errc(os.RemoveAll(stdoutFile.Name()))
	}
}

func InitExec(dirs *Dirs, path string, url string, what string) (string) {
	path = filepath.Join(dirs.Exe, path)

	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		log.Println("Installing", what)
		err = getter.GetFile(path, url)
	}

	if os.IsExist(err) {
		log.Println("Located executable:", path)
	}

	Errc(err)
	
	return path
}

func RobloxFind(dirs *Dirs, giveDir bool, exe string) (string) {
	var final string
	user := os.Getenv("USER")

	if user == "" {
		log.Fatal("Failed to get $USER variable")
	}

	var programDirs = []string {
		filepath.Join("users", user, "AppData/Local"),
		"Program Files (x86)",
	}

	for _, programDir := range programDirs {
		versionDir, err := os.Open(filepath.Join(dirs.Pfx, "drive_c", programDir, "Roblox/Versions"))

		if os.IsNotExist(err) {
			continue
		}

		versionDirs, err := versionDir.Readdir(0)
		Errc(err)

		for _, v := range versionDirs {
			checkExe, err := os.Open(filepath.Join(versionDir.Name(), v.Name(), exe))

			if err != nil {
				continue
			}
			defer checkExe.Close()

			if giveDir == false {	
				final = checkExe.Name()
			} else {
				final = filepath.Join(versionDir.Name(), v.Name())
			}

			break
		}
		defer versionDir.Close()
	}
	return final
}


func RbxFpsUnlocker(dirs *Dirs) {
	fpsUnlockerPath := InitExec(dirs, "rbxfpsunlocker.exe", FPSUNLOCKERURL, "FPS Unlocker")

	var settings = []string {
		"UnlockClient=true",
		"UnlockStudio=true",
		"FPSCapValues=[30.000000, 60.000000, 75.000000, 120.000000, 144.000000, 165.000000, 240.000000, 360.000000]",
		"FPSCapSelection=0",
		"FPSCap=0.000000",
		"CheckForUpdates=false",
		"NonBlockingErrors=true",
		"SilentErrors=true",
		"QuickStart=true",
	}

	settingsFile, err := os.Create(filepath.Join(dirs.Cache, "settings"))
	Errc(err)
	defer settingsFile.Close()

	// FIXME: compare settings file, to check if user has modified the settings file
	log.Println("Writing custom rbxfpsunlocker settings to", settingsFile.Name())
	for _, setting := range settings {
		_, err := fmt.Fprintln(settingsFile, setting + "\r")
		Errc(err)
	}

	log.Println("Launching FPS Unlocker")
	Exec(dirs, "wine", fpsUnlockerPath)
}

func RobloxLaunch(dirs *Dirs, exe string, url string, what string, arg string) {
	if RobloxFind(dirs, false, exe) == "" {
		installerPath := InitExec(dirs, exe, url, what)
		Exec(dirs, "wine", installerPath, "-fast")
		Errc(os.RemoveAll(installerPath))
	}

	log.Println("Launching Roblox Player")
	Exec(dirs, "wine", RobloxFind(dirs, false, exe), "-app", "-fast")
	//RbxFpsUnlocker(dirs)
}

func PfxKill(dirs *Dirs) {
	log.Println("Killing wineprefix")
	Exec(dirs, "wineserver", "-k")
}
