package bootstrapper

import (
	"fmt"
	"strconv"

	"golang.org/x/sync/errgroup"
)

type Package struct {
	Name     string
	Checksum string
	Size     int64
}

type Packages []Package

func ParsePackages(manif []string) (Packages, error) {
	pkgs := make(Packages, 0)

	if manif[0] != "v0" {
		return pkgs, fmt.Errorf("unhandled package manifest version: %s", manif)
	}

	for i := 1; i <= len(manif)-4; i += 4 {
		if IsExcluded(manif[i]) {
			continue
		}

		size, err := strconv.ParseInt(manif[i+2], 10, 64)
		if err != nil {
			return pkgs, err
		}

		pkgs = append(pkgs, Package{
			Name:     manif[i],
			Checksum: manif[i+1],
			Size:     size,
		})
	}

	return pkgs, nil
}

func (pkgs Packages) Perform(fn func(Package) error) error {
	var eg errgroup.Group

	for _, pkg := range pkgs {
		pkg := pkg

		eg.Go(func() error {
			return fn(pkg)
		})
	}

	return eg.Wait()
}
