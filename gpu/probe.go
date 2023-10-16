package gpu

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
)

var (
	cardPattern = regexp.MustCompile("card([0-9]+)$")
	edpPattern  = regexp.MustCompile("card([0-9]+)-eDP-[0-9]+$")
	drmPath     = "/sys/class/drm"
	driverPath  = "device/driver"
	vidPath     = "device/vendor"
	didPath     = "device/device"
)

// Note: sysfs is located entirely in memory, and as a result does not have IO errors.
// No error handling when calling IO operations is done.
func GetSystemGPUs() ([]*GPU, map[string]*GPU) {
	var gpus = make([]*GPU, 0)
	idDict := make(map[string]*GPU, 100)

	dirEntries, _ := os.ReadDir(drmPath)

	for _, v := range dirEntries {
		name := v.Name()
		submatch := cardPattern.FindStringSubmatch(name)
		eDPSubmatch := edpPattern.FindStringSubmatch(name)

		if submatch != nil {
			i, _ := strconv.Atoi(submatch[1])

			gpuPath := path.Join(drmPath, name)

			gpu := new(GPU)
			gpus = append(gpus, gpu)

			gpus[i].path = gpuPath

			driverPath, _ = filepath.EvalSymlinks(filepath.Join(gpu.path, "device/driver"))
			gpu.driver = driverPath

			vid, _ := os.ReadFile(path.Join(gpuPath, vidPath))
			did, _ := os.ReadFile(path.Join(gpuPath, didPath))

			gpu.SetId(vid, did)

			idDict[gpu.id] = gpus[i]

		} else if eDPSubmatch != nil {
			i, _ := strconv.Atoi(eDPSubmatch[0])
			gpus[i].eDP = true
		}
	}
	return gpus, idDict
}
