package bootstrapper

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/vinegarhq/vinegar/util"
)

func (p *Package) Verify(src string) error {
	log.Printf("Verifying Package %s (%s)", p.Name, p.Checksum)

	pkgFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer pkgFile.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, pkgFile); err != nil {
		return err
	}

	if p.Checksum != hex.EncodeToString(hash.Sum(nil)) {
		return fmt.Errorf("package %s (%s) is corrupted", p.Name, src)
	}

	return nil
}

func (p *Package) Download(dest, deployURL string) error {
	if err := p.Verify(dest); err == nil {
		log.Printf("Package %s is already downloaded", p.Name)
		return nil
	}

	log.Printf("Downloading Package %s (%s)", p.Name, dest)

	if err := util.Download(deployURL+"-"+p.Name, dest); err != nil {
		return fmt.Errorf("download package %s: %w", p.Name, dest, err)
	}

	return p.Verify(dest)
}

func (p *Package) Fetch(dest, deployURL string) error {
	err := p.Download(dest, deployURL)
	if err == nil {
		return nil
	}

	log.Printf("Failed to fetch package %s: %s, retrying...", p.Name, errors.Unwrap(err))

	return p.Download(dest, deployURL)
}

func (p *Package) Extract(src, dest string) error {
	if err := extract(src, dest); err != nil {
		return fmt.Errorf("extract package %s (%s): %w", p.Name, src, err)
	}

	log.Printf("Extracted Package %s (%s)", p.Name, p.Checksum)
	return nil
}
