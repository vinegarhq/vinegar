//go:build freebsd

package sysinfo

import (
	"syscall"
)

func getCPUName() string {
	if model, err := syscall.Sysctl("hw.model"); err == nil {
		return model
	}

	return ""
}
