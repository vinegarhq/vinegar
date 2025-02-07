//go:build freebsd

package sysinfo

import (
	"syscall"
)

func cpuName() string {
	if model, err := syscall.Sysctl("hw.model"); err == nil {
		return model
	}

	return ""
}
