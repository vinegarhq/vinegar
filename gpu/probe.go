package gpu

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
)

var (
	cardPattern    = regexp.MustCompile("card([0-9]+)$")
	edpPattern     = regexp.MustCompile("card([0-9]+)-eDP-[0-9]+$")
	drmPath        = "/sys/class/drm"
	driverLinkPath = "device/driver"
)

var gpus = make([]*GPU, 0)

// Note: sysfs is located entirely in memory, and as a result does not have IO errors.
// No error handling when calling IO operations is done.
func GetSystemGPUs() []*GPU {
	if len(gpus) > 0 {
		return gpus
	}

	dirEntries, _ := os.ReadDir(drmPath)

	for _, v := range dirEntries {
		name := v.Name()
		submatch := cardPattern.FindStringSubmatch(name)
		eDPSubmatch := edpPattern.FindStringSubmatch(name)

		if eDPSubmatch != nil {
			i, _ := strconv.Atoi(eDPSubmatch[1])
			gpus[i].eDP = true
		}
		if submatch == nil {
			continue
		}

		i, _ := strconv.Atoi(submatch[1])

		gpu := GPU{}
		gpus = append(gpus, &gpu)
		gpus[i].path = path.Join(drmPath, name)
		driverPath, _ := filepath.EvalSymlinks(filepath.Join(gpus[i].path, driverLinkPath))
		gpus[i].driver = driverPath
	}

	return gpus
}
