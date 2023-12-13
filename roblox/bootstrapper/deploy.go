package bootstrapper

import (
	"errors"
	"fmt"
	"log"

	"github.com/vinegarhq/vinegar/util"
)

// Verify checks the named package source file against it's checksum
func (p *Package) Verify(src string) error {
	log.Printf("Verifying Package %s (%s)", p.Name, p.Checksum)

	if err := util.VerifyFileMD5(src, p.Checksum); err != nil {
		return fmt.Errorf("verify package %s: %w", p.Name, err)
	}

	return nil
}

// Download will download the package to the named dest destination
// directory with the given deployURL deploy mirror; if the package
// exists and has the correct checksum, it will return immediately.
func (p *Package) Download(dest, deployURL string) error {
	if err := p.Verify(dest); err == nil {
		log.Printf("Package %s is already downloaded", p.Name)
		return nil
	}

	log.Printf("Downloading Package %s (%s)", p.Name, dest)

	if err := util.Download(deployURL+"-"+p.Name, dest); err != nil {
		return fmt.Errorf("download package %s (%s): %w", p.Name, dest, err)
	}

	return p.Verify(dest)
}

// Fetch is a wrapper around Download to account for failures.
func (p *Package) Fetch(dest, deployURL string) error {
	err := p.Download(dest, deployURL)
	if err == nil {
		return nil
	}

	log.Printf("Failed to fetch package %s: %s, retrying...", p.Name, errors.Unwrap(err))

	return p.Download(dest, deployURL)
}

// Extract extracts the named package source file to a given destination directory
func (p *Package) Extract(src, dest string) error {
	if err := extract(src, dest); err != nil {
		return fmt.Errorf("extract package %s (%s): %w", p.Name, src, err)
	}

	log.Printf("Extracted Package %s (%s)", p.Name, p.Checksum)
	return nil
}
