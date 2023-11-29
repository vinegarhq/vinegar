//go:build linux

package sysinfo

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

func cpuModel() (string, bool) {
	column := regexp.MustCompile("\t+: ")

	f, _ := os.Open("/proc/cpuinfo")
	defer f.Close()

	s := bufio.NewScanner(f)

	n := "unknown cpu"
	l := false

	for s.Scan() {
		sl := column.Split(s.Text(), 2)
		if sl == nil {
			continue
		}

		// pfft, who needs multiple cpus? just return if we got all we need
		if sl[0] == "model name" {
			n = sl[1]
		}

		if sl[0] == "flags" {
			l = strings.Contains(sl[1], "split_lock_detect")
			break
		}

	}

	return n, l
}
