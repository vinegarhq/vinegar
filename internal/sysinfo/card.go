package sysinfo

import (
	"path"
)

// Card is a representation of a system GPU
type Card struct {
	Index     int    // Internal Kernel index
	Path      string // Path to the drm card
	Device    string // Path to the PCI device
	Driver    string // Base driver name
	Embedded  bool   // Integrated display
	VendorID  string
	ProductID string

	// Metadata added in top-level implementation
	Vendor  string
	Product string
}

func (c Card) String() string {
	return c.Vendor + " " + c.Product
}

func (c *Card) Addr() string {
	return path.Base(c.Device)
}
