//go:build linux

package sysinfo

import (
	"strings"
	"syscall"
)

type kernel struct {
	Release string
	Version string
}

func getKernel() kernel {
	var un syscall.Utsname
	_ = syscall.Uname(&un)

	return kernel{
		Release: unameString(un.Release),
		Version: unameString(un.Version),
	}
}

func (k kernel) String() string {
	return k.Release + " " + k.Version
}

func unameString(unarr [65]int8) string {
	var sb strings.Builder
	for _, b := range unarr[:] {
		if b == 0 {
			break
		}
		sb.WriteByte(byte(b))
	}
	return sb.String()
}
