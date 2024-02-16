// Package netutil provides shared utility networking functions.
package netutil

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// DrawFunc is the callback type for drawing progress, it will
// be ran in a goroutine.
type DrawFunc func(float32)

type progressCounter struct {
	total   uint64
	current uint64
	draw    DrawFunc
}

func (pc *progressCounter) Write(p []byte) (int, error) {
	n := len(p)
	pc.current += uint64(n)
	go pc.draw(float32(pc.current) / float32(pc.total))
	return n, nil
}

// ErrBadStatus is the error returned by Download and Body
// if the returned HTTP status code is not http.StatusOK.
var ErrBadStatus = errors.New("bad status")

// DownloadProgress downloads the named url to the named file, using
// df as the callback for progress. No retry will be checked here.
func DownloadProgress(url, file string, df DrawFunc) error {
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
		draw:  df,
	}

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

// Body retrieves the body of the named url to string form.
func Body(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %s", ErrBadStatus, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
