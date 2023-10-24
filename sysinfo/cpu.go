//go:build linux

package sysinfo

import (
	"bufio"
	"os"
	"regexp"
)

func cpuModel() string {
	column := regexp.MustCompile("\t+: ")

	f, _ := os.Open("/proc/cpuinfo")
	defer f.Close()

	s := bufio.NewScanner(f)

	for s.Scan() {
		sl := column.Split(s.Text(), 2)
		if sl == nil {
			continue
		}

		// pfft, who needs multiple cpus? just return if we got all we need
		if sl[0] == "model name" {
			return sl[1]
		}
	}

	return "unknown cpu"
}
