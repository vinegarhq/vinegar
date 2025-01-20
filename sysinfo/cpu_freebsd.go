//go:build freebsd

package sysinfo

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	cpulib "golang.org/x/sys/cpu"
)

func getCPU() Processor {
	c := Processor{
		Name: "unknown cpu",
		AVX:  cpulib.X86.HasAVX,
	}

	column := regexp.MustCompile("\t+: ")

	f, _ := os.Open("/linproc/cpuinfo")
	defer f.Close()

	s := bufio.NewScanner(f)

	for s.Scan() {
		sl := column.Split(s.Text(), 2)
		if sl == nil {
			continue
		}

		// pfft, who needs multiple cpus? just return if we got all we need
		if sl[0] == "model name" {
			c.Name = sl[1]
		}

		if sl[0] == "flags" {
			c.SplitLockDetect = strings.Contains(sl[1], "split_lock_detect")
			break
		}

	}

	return c
}
