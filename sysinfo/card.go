package sysinfo

// Card is a representation of a system GPU
type Card struct {
	Path     string // Path to the drm card
	Device   string // Path to the PCI device
	Driver   string // Base driver name
	Embedded bool   // Integrated display
}
