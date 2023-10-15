package api

import (
	"fmt"
)

type Thumbnail struct {
	TargetID int64  `json:"targetId"`
	State    string `json:"state"`
	ImageURL string `json:"imageUrl"`
	Version  string `json:"version"`
}

type ThumbnailResponse struct {
	Data []Thumbnail `json:"data"`
}

func GetGameIcon(universeID, returnPolicy, size, format string, isCircular bool) (Thumbnail, error) {
	var tnr ThumbnailResponse
	err := UnmarshalBody(
		fmt.Sprintf("https://thumbnails.roblox.com/v1/games/icons?universeIds=%s&returnPolicy=%s&size=%s&format=%s&isCircular=%t",
			universeID, returnPolicy, size, format, isCircular), &tnr,
	)
	if err != nil {
		return Thumbnail{}, err
	}

	return tnr.Data[0], nil
}
