package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const (
	RCOURL = "https://raw.githubusercontent.com/L8X/Roblox-Client-Optimizer/main/ClientAppSettings.json"
)

func RobloxFind(giveDir bool, exe string) string {
	for _, programDir := range programDirs {
		versionDir := filepath.Join(programDir, "Roblox/Versions")

		rootExe := filepath.Join(versionDir, exe)
		if _, e := os.Stat(rootExe); e == nil {
			if !giveDir {
				return rootExe
			}

			return versionDir
		}

		versionExe, _ := filepath.Glob(filepath.Join(versionDir, "*", exe))

		if versionExe == nil {
			continue
		}

		if !giveDir {
			return versionExe[0]
		}

		return filepath.Dir(versionExe[0])
	}

	return ""
}

func RobloxInstall(url string) error {
	log.Println("Installing Roblox")

	installerPath := filepath.Join(Dirs.Cache, "rbxinstall.exe")

	if err := Download(url, installerPath); err != nil {
		return err
	}

	if err := Exec("wine", true, installerPath); err != nil {
		return err
	}

	if err := os.RemoveAll(installerPath); err != nil {
		return err
	}

	return nil
}

func RobloxSetRenderer(renderer string) {
	possibleRenderers := []string{
		"OpenGL",
		"D3D11FL10",
		"D3D11",
		"Vulkan",
	}

	validRenderer := false

	for _, r := range possibleRenderers {
		if renderer == r {
			validRenderer = true
		}
	}

	if !validRenderer {
		log.Fatal("invalid renderer, must be one of:", possibleRenderers)
	}

	for _, r := range possibleRenderers {
		isRenderer := r == renderer
		Config.FFlags["FFlagDebugGraphicsPrefer"+r] = isRenderer
		Config.FFlags["FFlagDebugGraphicsDisable"+r] = !isRenderer
	}
}

func RobloxApplyFFlags(app string, dir string) error {
	fflags := make(map[string]interface{})

	if app == "Studio" {
		return nil
	}

	fflagsDir := filepath.Join(dir, app+"Settings")
	CheckDirs(DirMode, fflagsDir)

	fflagsFile, err := os.Create(filepath.Join(fflagsDir, app+"AppSettings.json"))
	if err != nil {
		return err
	}

	if Config.ApplyRCO {
		log.Println("Applying RCO FFlags")

		if err := Download(RCOURL, fflagsFile.Name()); err != nil {
			return err
		}
	}

	RobloxSetRenderer(Config.Renderer)

	fflagsFileContents, err := os.ReadFile(fflagsFile.Name())
	if err != nil {
		return err
	}

	if err := json.Unmarshal(fflagsFileContents, &fflags); err != nil {
		return err
	}

	log.Println("Applying custom FFlags")

	for fflag, value := range Config.FFlags {
		fflags[fflag] = value
	}

	fflagsJSON, err := json.MarshalIndent(fflags, "", "  ")
	if err != nil {
		return err
	}

	if _, err := fflagsFile.Write(fflagsJSON); err != nil {
		return err
	}

	return nil
}

func RobloxLaunch(exe string, app string, args ...string) {
	EdgeDirSet(DirROMode, true)

	if RobloxFind(false, exe) == "" {
		if err := RobloxInstall("https://www.roblox.com/download/client"); err != nil {
			log.Fatal("failed to install roblox:", err)
		}
	}

	robloxRoot := RobloxFind(true, exe)

	if robloxRoot == "" {
		log.Fatal("failed to find roblox")
	}

	if err := RobloxApplyFFlags(app, robloxRoot); err != nil {
		log.Fatal("failed to apply fflags:", err)
	}

	DxvkStrap()

	log.Println("Launching", exe)
	args = append([]string{filepath.Join(robloxRoot, exe)}, args...)

	prog := "wine"

	if Config.GameMode {
		args = append([]string{"wine"}, args...)
		prog = "gamemoderun"
	}

	if err := Exec(prog, true, args...); err != nil {
		log.Fatal("roblox exec err: ", err)
	}

	CommLoop(exe[:15])
}
