package bootstrapper

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/vinegarhq/vinegar/util"
)

// Package is a representation of a Binary package.
type Package struct {
	Name     string
	Checksum string
	Size     int64
	ZipSize  int64
}

type Packages []Package

// Verify checks the named package source file against it's checksum
func (p *Package) Verify(src string) error {
	log.Printf("Verifying Package %s (%s)", p.Name, p.Checksum)

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	fsum := hex.EncodeToString(h.Sum(nil))

	if p.Checksum != fsum {
		return fmt.Errorf("package %s (%s) is corrupted, please re-download or delete package", p.Name, src)
	}

	return nil
}

// Download will download the package to the named dest destination
// directory with the given deployURL deploy mirror; if the package
// exists and has the correct checksum, it will return immediately.
//
// If downloading the package fails, it will attempt to re-download
// only once.
func (p *Package) Download(dest, deployURL string) error {
	if err := p.Verify(dest); err == nil {
		log.Printf("Package %s is already downloaded", p.Name)
		return nil
	}

	log.Printf("Downloading Package %s (%s)", p.Name, dest)

	if err := util.Download(deployURL+"-"+p.Name, dest); err == nil {
		return p.Verify(dest)
	}

	log.Printf("Failed to fetch package %s, retrying...", p.Name)

	if err := util.Download(deployURL+"-"+p.Name, dest); err != nil {
		return fmt.Errorf("download package %s: %w", p.Name, err)
	}

	return p.Verify(dest)
}

// Extract extracts the named package source file to a given destination directory
func (p *Package) Extract(src, dest string) error {
	if err := extract(src, dest); err != nil {
		return fmt.Errorf("extract package %s (%s): %w", p.Name, src, err)
	}

	log.Printf("Extracted Package %s (%s): %s", p.Name, p.Checksum, dest)
	return nil
}

func verifyFileMD5(name string, sum string) error {

	return nil
}
