package sysinfo

import (
	"strings"

	"golang.org/x/sys/unix"
)

func getKernel() string {
	var un unix.Utsname
	_ = unix.Uname(&un)

	var sb strings.Builder
	for _, b := range un.Release[:] {
		if b == 0 {
			break
		}
		sb.WriteByte(byte(b))
	}
	return sb.String()
}
