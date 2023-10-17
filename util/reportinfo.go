/*
This file is for detecting the user's system for helpdesk purposes.
This is only a prototype.

We will collect CPU flags and check for AVX, distro, kernel, flatpak, and config.
More features may be added at a later time...

Wael, if you would like to combine the file reading into a single function, that would be cool.
 - lunarlattice 10/16/23
*/

package util

import (
	"github.com/vinegarhq/vinegar/internal/config"
	"os"
	"errors"
	"regexp"
)

type SysInfo struct {
	AVXAvailable    bool
	Distro		string //Done
	Kernel		string // Done
	InFlatpak	bool // Done
	Config		string // Done
}
func GenerateInfo(currentConfigurationPath string) (SysInfo, error){
	//TODO, returns struct Sysinfo.
	currentSystem := &SysInfo{}

	// Check for AVX
	if cpufile, err := os.ReadFile("/proc/cpuinfo"); err != nil {
		return SysInfo{}, err
	} else {
		exp := regexp.MustCompile(`avx`)
		matches := exp.FindStringSubmatch(string(cpufile))
		currentSystem.AVXAvailable = (len(matches) > 0)
	}


	// Get Distro
	if distro, err := os.ReadFile("/etc/os-release"); err != nil {
		return SysInfo{}, err
	} else {
		currentSystem.Distro = string(distro)
	}

	// Get kernel
	if kernel, err := os.ReadFile("/proc/version"); err != nil {
		return SysInfo{}, err
	} else {
		currentSystem.Kernel = string(kernel)
	}

	// Read the config and store
	if config, err := os.ReadFile(currentConfigurationPath); err != nil {
		return SysInfo, err
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