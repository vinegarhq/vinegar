//go:build dragonfly || freebsd || netbsd

package util

import (
	"os"
	"path/filepath"
	"strings"
)

func CommFound(query string) bool {
	// 'cmdline' is used because FreeBSD and NetBSD procfs
	// is not the same as Linux and does not have 'comm'.
	// Additionally, procfs needs to be enabled to work.
	comms, _ := filepath.Glob("/proc/*/cmdline")

	for _, comm := range comms {
		c, err := os.ReadFile(comm)

		if err == nil && strings.Contains(string(c), query) {
			return true
		}
	}

	return false
}
