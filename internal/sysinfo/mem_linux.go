//go:build linux

package sysinfo

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func getTotalMemory() uint64 {
	f, _ := os.Open("/proc/meminfo")
	defer f.Close()

	s := bufio.NewScanner(f)

	for s.Scan() {
		parts := strings.Split(s.Text(), ":")
		if parts[0] != "MemTotal" {
			continue

		}

		value, err := strconv.ParseUint(
			strings.TrimSuffix(strings.TrimSpace(parts[1]), " kB"),
			10, 64)
		if err != nil {
			break
		}
		return value * 1024
	}

	return 64
}
