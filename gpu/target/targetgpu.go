package target

import (
	"strconv"
)

type PrimeType int

const (
	PrimeIntegrated PrimeType = iota
	PrimeDiscrete
	PrimeNone
	PrimeUnknown
)

type TargetGPU struct {
	Id    int
	Prime bool
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

// Select a target GPU based on intended PRIME configuration
func (t *TargetGPU) SetPrimeTarget(p PrimeType) {
	switch p {
	case PrimeIntegrated:
		t.Id = 0
	case PrimeDiscrete:
		t.Id = 1
	case PrimeNone:
		t.Id = -1 //Ignore prime
	}
	t.Prime = true
}

// Select the target GPU with given index
func (t *TargetGPU) SetDirectTarget(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	t.Id = i
	return nil
}

func (t *TargetGPU) SetTarget(s string) error {
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

func New() TargetGPU {
	return TargetGPU{}
}
