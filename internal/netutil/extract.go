package netutil

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

// ExtractURL will decompress the given XZ compressed tarball URL
// into path.
func ExtractURL(url string, dir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()

	xz, err := xz.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("xz: %w", err)
	}
	r := tar.NewReader(xz)

	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}

		dest := filepath.Join(dir, h.Name)

		if !strings.HasPrefix(dest, filepath.Clean(dir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal package file path: %s", dest)
		}

		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dest, os.FileMode(h.Mode)); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
			continue
		case tar.TypeSymlink:
			if err := os.Symlink(h.Linkname, dest); err != nil {
				return err
			}
			continue
		case tar.TypeLink:
			if err := os.Link(h.Linkname, dest); err != nil {
				return err
			}
			continue
		case tar.TypeReg:
		default:
			continue
		}

		err = func() error {
			f, err := os.OpenFile(dest,
				os.O_WRONLY|os.O_CREATE, h.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("create: %w", err)
			}
			defer f.Close()

			if err := os.Chtimes(dest, h.AccessTime, h.ModTime); err != nil {
				return err
			}

			_, err = io.Copy(f, r)
			if err != nil {
				return fmt.Errorf("copy: %w", err)
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}
