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

func SetClient(client *http.Client) {
	httpClient = client
}

func Request(method, service, endpoint string, v interface{}) error {
	log.Printf("Performing %s request on %s/%s", method, service, endpoint)

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

	if resp.StatusCode >= 400 {
		errsResp := new(errorsResponse)
		if err := json.NewDecoder(resp.Body).Decode(errsResp); err != nil {
			return err
		}

		return errsResp
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}
