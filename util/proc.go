package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func CommFound(query string) bool {
	var comms, _ []string
	switch runtime.GOOS {
	case "linux":
		comms, _ = filepath.Glob("/proc/*/comm")
	default:
		// FreeBSD, NetBSD, etc..
		// Enabled procfs is required.
		comms, _ = filepath.Glob("/proc/*/cmdline")
	}

	for _, comm := range comms {
		c, err := os.ReadFile(comm)
		// On Linux, the 'comm' file contains a newline, we remove it, as it
		// will mess up the query. Hence 'minus'ing the length by 1 removes
		// the newline.
		if runtime.GOOS == "linux" && err == nil &&
		strings.Contains(string(c)[:len(c)-1], query) {
			return true
		// On FreeBSD and NetBSD, this issue doesn't exist. Subtracting length
		// by one doesn't work as 'cmdline' doesn't end in a newline character.
		} else if err == nil && strings.Contains(string(c), query) {
			return true
		}
	}

	return false
}
