// Package api provides API routines to interact with Roblox's web API.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

const APIURL = "https://%s.roblox.com/%s"

var httpClient = &http.Client{}

var ErrBadStatus = errors.New("bad status")

// SetClient sets the http.Client used to make API requests.
func SetClient(client *http.Client) {
	httpClient = client
}

// Request makes a API request given method, service, endpoint, and data
// to send to the endpoint with the given method.
func Request(method, service, endpoint string, v interface{}) error {
	log.Printf("Performing Roblox API %s %s request on %s", method, service, endpoint)

	url := fmt.Sprintf(APIURL, service, endpoint)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Return the given API error only if the decoder succeeded
		errsResp := new(errorsResponse)
		if err := json.NewDecoder(resp.Body).Decode(errsResp); err == nil {
			return errsResp
		}

		return fmt.Errorf("%w: %s", ErrBadStatus, resp.Status)
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}
