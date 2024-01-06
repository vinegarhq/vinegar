package sysinfo

// Processor is a representation of a system CPU
type Processor struct {
	Name            string
	AVX             bool
	SplitLockDetect bool
}
