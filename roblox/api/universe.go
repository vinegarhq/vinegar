package api

import (
	"strconv"
)

type universeIdResponse struct {
	UniverseID int64 `json:"universeId"`
}

// GetUniverseID uses an undocumented Roblox web API service to retrieve
// the universeID for the named placeID.
func GetUniverseID(placeID string) (string, error) {
	var uidr universeIdResponse

	// This API is undocumented.
	err := Request("GET", "apis", "universes/v1/places/"+placeID+"/universe", &uidr)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(uidr.UniverseID, 10), nil
}
