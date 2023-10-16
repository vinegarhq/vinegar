/*
This file is for detecting the user's system for helpdesk purposes.
This is only a prototype.

We will collect CPU flags and check for AVX, distro, kernel, flatpak, and config.
More features may be added at a later time...

 - lunarlattice 10/16/23
*/

package util

import (
	"github.com/vinegarhq/vinegar/internal/config"
	"os"
	"errors"
)

type SysInfo struct {
	CPUFlags	string
	AVXAvailable    bool
	Distro		string
	Kernel		string
	InFlatpak	bool // Done
	Config		string // Done
}
func GenerateInfo(currentConfigurationPath string) (SysInfo, error){
	//TODO, returns struct Sysinfo.
	currentSystem := &SysInfo{}

	// Read the config and store
	if config, err := os.ReadFile(currentConfigurationPath); err != nil {
		return *currentSystem, err
	} else {
		currentSystem.Config = string(config)
	}

	// Check if in flatpak
	if _, err := os.Stat("/.flatpak-info"); err == nil {
		currentSystem.InFlatpak = true
	} else if errors.Is(err, os.ErrNotExist) {
		currentSystem.InFlatpak = false
	}

	return *currentSystem, nil
}