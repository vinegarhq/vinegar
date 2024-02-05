package bootstrapper

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/vinegarhq/vinegar/internal/netutil"
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
	slog.Info("Verifying Package", "name", p.Name, "path", src)

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
		return fmt.Errorf("package %s is corrupted, please re-download or delete package", p.Name)
	}

	return nil
}

// Download will download the package to the named dest destination
// directory with the given deployURL deploy mirror; if the package
// exists and has the correct checksum, it will return immediately.
func (p *Package) Download(dest, deployURL string) error {
	if err := p.Verify(dest); err == nil {
		slog.Info("Package is already downloaded", "name", p.Name, "file", dest)
		return nil
	}

	url := deployURL + "-" + p.Name
	slog.Info("Downloading package", "url", url, "path", dest)

	if err := netutil.Download(url, dest); err != nil {
		return fmt.Errorf("download package %s: %w", p.Name, err)
	}

	return p.Verify(dest)
}

// Extract extracts the named package source file to a given destination directory
func (p *Package) Extract(src, dest string) error {
	if err := extract(src, dest); err != nil {
		return fmt.Errorf("extract package %s (%s): %w", p.Name, src, err)
	}

	slog.Info("Extracted package", "name", p.Name, "path", src, "dest", dest)
	return nil
}
