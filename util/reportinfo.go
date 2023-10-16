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
)

type SysInfo struct {
	CPUFlags	string
	AVXAvailable    bool
	Distro		string
	Kernel		string
	InFlatpak	bool
	Config		config.Config
}
func GenerateInfo(currentConfigurationPath string) (SysInfo){
	//TODO, returns struct Sysinfo.
	return string
}