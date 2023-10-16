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

var gpuCache *[]*GPU

// Note: sysfs is located entirely in memory, and as a result does not have IO errors.
// No error handling when calling IO operations is done.
func GetSystemGPUs() []*GPU {
	if gpuCache != nil {
		return *gpuCache
	}

	var gpus = make([]*GPU, 0)
	dirEntries, _ := os.ReadDir(drmPath)

	for _, v := range dirEntries {
		name := v.Name()
		submatch := cardPattern.FindStringSubmatch(name)
		eDPSubmatch := edpPattern.FindStringSubmatch(name)

		if eDPSubmatch != nil {
			i, _ := strconv.Atoi(eDPSubmatch[0])
			gpus[i].eDP = true
		}
		if submatch != nil {
			continue
		}

		i, _ := strconv.Atoi(submatch[1])

		gpuPath := path.Join(drmPath, name)
		gpu := new(GPU)
		gpus = append(gpus, gpu)
		gpus[i].path = gpuPath
		driverPath, _ := filepath.EvalSymlinks(filepath.Join(gpu.path, driverLinkPath))
		gpu.driver = driverPath
	}
	gpuCache = &gpus
	return gpus
}
