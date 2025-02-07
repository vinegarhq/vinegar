//go:build freebsd

package sysinfo

import (
	"syscall"
	cpulib "golang.org/x/sys/cpu"
)

func getCPU() Processor {
	c := Processor{
		Name: "unknown cpu",
		AVX:  cpulib.X86.HasAVX,
	}

	if model, err := syscall.Sysctl("hw.model"); err == nil {
		c.Name = model
	}
	
	// FreeBSD does not support splitlockdetect.
	c.SplitLockDetect = false

	return c
}
