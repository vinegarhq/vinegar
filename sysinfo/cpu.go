//go:build linux

package sysinfo

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	cpulib "golang.org/x/sys/cpu"
)

type cpu struct {
	Name            string
	AVX             bool
	SplitLockDetect bool
}

func (c cpu) String() string {
	return fmt.Sprintf(`%s
  * Supports AVX: %t
  * Supports split lock detection: %t`, c.Name, c.AVX, c.SplitLockDetect)
}

func getCPU() cpu {
	c := cpu{
		Name: "unknown cpu",
		AVX: cpulib.X86.HasAVX,
	}

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
			c.Name = sl[1]
		}

		if sl[0] == "flags" {
			c.SplitLockDetect = strings.Contains(sl[1], "split_lock_detect")
			break
		}

	}

	return c
}
