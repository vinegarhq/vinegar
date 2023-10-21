package api

import (
	"strconv"
)

type UniverseIdResponse struct {
	UniverseID int64 `json:"universeId"`
}

func GetUniverseID(placeID string) (string, error) {
	var uidr UniverseIdResponse

	// This API is undocumented.
	err := Request("GET", "apis", "universes/v1/places/"+placeID+"/universe", &uidr)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(uidr.UniverseID, 10), nil
}
