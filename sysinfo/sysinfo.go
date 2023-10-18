package sysinfo

import (
	"os"
	"io/fs"
	"strconv"
	"bufio"
	"regexp"
	"syscall"
	"strings"
	"path/filepath"
)

type Kernel struct {
	Release string
	Version string
}

type CPU struct {
	Model string
	Flags []string
}

type GPU struct {
	Path       string
	Integrated bool
	Index      int
	Driver     string
}

type GPUs []GPU

func GetKernel() Kernel {
	var un syscall.Utsname
	_ = syscall.Uname(&un)

	unameString := func(unarr [65]int8) string {
		var sb strings.Builder
		for _, b := range unarr[:] {
			if b == 0 {
				break
			}
			sb.WriteByte(byte(b))
		}
		return sb.String()
	}

	return Kernel{
		Release: unameString(un.Release),
		Version: unameString(un.Version),
	}
}

func (k Kernel) String() string {
	return k.Release + " " + k.Version
}

func GetCPU() (cpu CPU) {
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
		if cpu.Model != "" && cpu.Flags != nil {
			break
		}

		switch sl[0] {
		case "model name":
			cpu.Model = sl[1]
		case "flags":
			cpu.Flags = strings.Split(sl[1], " ")
		}
	}

	if s.Err() != nil {
		return
	}

	return
}

func (cpu *CPU) String() string {
	return cpu.Model
}

func GetGPUs() (gpus GPUs) {
	card := regexp.MustCompile(`card([0-9]+)(?:-eDP-\d+)?$`)

	filepath.Walk("/sys/class/drm", func(p string, i fs.FileInfo, err error) error {
		var gpu GPU

		match := card.FindStringSubmatch(p)
		if match == nil {
			return nil
		}

		if len(match) == 2 {
			gpu.Integrated = true
		}

		gpu.Index, _ = strconv.Atoi(match[1])
		gpu.Driver, _ = filepath.EvalSymlinks(filepath.Join(p, "device/driver"))
		gpu.Path = p

		gpus = append(gpus, gpu)
		return nil
	})

	return
}

func InFlatpak() bool {
	_, err := os.Stat("/.flatpak-info")
	return err == nil
}
