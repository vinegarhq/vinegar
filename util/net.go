package util

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ErrBadStatus is the error returned by Download and Body
// if the returned HTTP status code is not http.StatusOK.
var ErrBadStatus = errors.New("bad status")

// Download downloads the named url to the named file.
func Download(url, file string) error {
	out, err := os.Create(file)
	if err != nil {
		return err
	}

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
