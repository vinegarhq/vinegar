// Package netutil provides shared utility networking functions.
package netutil

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type progressCounter struct {
	total   uint64
	current uint64
	pbar    *gtk.ProgressBar
}

func (pc *progressCounter) Write(p []byte) (int, error) {
	n := len(p)
	pc.current += uint64(n)
	return n, nil
}

// ErrBadStatus is the error returned by Download and Body
// if the returned HTTP status code is not http.StatusOK.
var ErrBadStatus = errors.New("bad status")

// DownloadProgress downloads the named url to the named file, using
// df as the callback for progress. No retry will be checked here.
func DownloadProgress(url, file string, pbar *gtk.ProgressBar) error {
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrBadStatus, resp.Status)
	}

	pc := &progressCounter{
		total: uint64(resp.ContentLength),
		pbar:  pbar,
	}

	var idlecb glib.SourceFunc = func(uintptr) bool {
		pbar.SetFraction(float64(pc.current) / float64(pc.total))
		return pc.current != pc.total
	}
	glib.TimeoutAdd(16, &idlecb, uintptr(unsafe.Pointer(nil)))

	_, err = io.Copy(out, io.TeeReader(resp.Body, pc))
	if err != nil {
		return err
	}

	return nil
}

// Download downloads the named url to the named file. If an error
// occurs when downloading the file. Download will retry 3 times before
// returning a final error.
func Download(url, file string) error {
	retries := 3
	for i := 0; i < retries; i++ {
		err := download(url, file)
		if err == nil {
			break
		}

		// additional condition for if the error was a file error or status error
		if _, ok := err.(*os.PathError); err != nil &&
			(i == retries-1 || ok || errors.Is(err, ErrBadStatus)) {
			os.Remove(file) // just remove the thing anyway on failure
			return err
		}

		log.Printf("Download %s failed, retrying...", url)
	}

	return nil
}

func download(url, file string) error {
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrBadStatus, resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
