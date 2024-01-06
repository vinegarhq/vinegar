//go:build linux

package sysinfo

import (
	"strings"
	"syscall"
)

func getKernel() string {
	var un syscall.Utsname
	_ = syscall.Uname(&un)

	var sb strings.Builder
	for _, b := range un.Release[:] {
		if b == 0 {
			break
		}
		sb.WriteByte(byte(b))
	}
	return sb.String()
}
