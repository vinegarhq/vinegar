package target

import (
	"strconv"
	"strings"
)

type PrimeType int

const (
	PrimeIntegrated PrimeType = iota
	PrimeDiscrete
	PrimeNone
	PrimeUnknown
)

type TargetGpu struct {
	Id      string
	Prime   bool
	IsIndex bool
}

func (pt PrimeType) String() string {
	switch pt {
	case PrimeIntegrated:
		return "integrated"
	case PrimeDiscrete:
		return "prime-discrete"
	case PrimeNone:
		return "none"
	default:
		return "unknown"
	}
}

func GetPrimeType(s string) PrimeType {
	if s == "" {
		s = "none"
	}

	switch s {
	case PrimeIntegrated.String():
		return PrimeIntegrated
	case PrimeDiscrete.String():
		return PrimeDiscrete
	case PrimeNone.String():
		return PrimeNone
	default:
		return PrimeUnknown
	}
}

func SanitizeGpuId(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "0x", "")
	return s
}

// Select a target GPU based on intended PRIME configuration
func (t *TargetGpu) SetPrimeTarget(p PrimeType) {
	switch p {
	case PrimeIntegrated:
		t.Id = "0"
	case PrimeDiscrete:
		t.Id = "1"
	case PrimeNone:
		t.Id = "" //Ignore prime
	}
	t.Prime = true
	t.IsIndex = true
}

// Select the target GPU with its index or VID:NID
func (t *TargetGpu) SetDirectTarget(s string) error {
	if strings.Contains(s, ":") { //vid:nid
		t.Id = SanitizeGpuId(s)
	} else { //index
		_, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		t.Id = s
		t.IsIndex = true
	}
	return nil
}

func (t *TargetGpu) SetTarget(s string) error {
	prime := GetPrimeType(s)

	if prime != PrimeUnknown {
		t.SetPrimeTarget(prime)
	} else {
		err := t.SetDirectTarget(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func New() TargetGpu {
	return TargetGpu{
		Id:      "",
		Prime:   false,
		IsIndex: false,
	}
}
