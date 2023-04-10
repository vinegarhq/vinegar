package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type ClientVersion struct {
	Version             string `json:"version"`
	VersionGUID         string `json:"clientVersionUpload"`
	BootstrapperVersion string `json:"bootstrapperVersion"`
}

var client http.Client

func RobloxAPIRequest(api, endpoint, method string, json []byte) []byte {
	url := "https://" + api + ".roblox.com/" + endpoint
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal("failed to perform roblox api request: ", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	resp.Body.Close()

	return body
}

func RobloxPlayerLatestVersion() string {
	var versionInfo ClientVersion

	resp := RobloxAPIRequest("clientsettings", "v2/client-version/WindowsPlayer/channel/LIVE", "GET", nil)

	if err := json.Unmarshal(resp, &versionInfo); err != nil {
		log.Fatal(err)
	}

	return versionInfo.VersionGUID
}
