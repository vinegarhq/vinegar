package api

import (
	"fmt"
)

// Thumbnail is a representation of the Roblox ThumbnailResponse model.
type Thumbnail struct {
	TargetID int64  `json:"targetId"`
	State    string `json:"state"`
	ImageURL string `json:"imageUrl"`
	Version  string `json:"version"`
}

// Thumbnail is a representation of the Roblox ApiArrayResponse ThumbnailResponse model.
type thumbnailResponse struct {
	Data []Thumbnail `json:"data"`
}

// GetGameIcon gets the thumbnail URL for the given universeID, refer to the
// [Thumbnails API documentation] for more information.
//
// [[Thumbnails API documentation]: https://thumbnails.roblox.com/docs/index.html
func GetGameIcon(universeID, returnPolicy, size, format string, isCircular bool) (Thumbnail, error) {
	var tnr thumbnailResponse

	err := Request("GET", "thumbnails",
		fmt.Sprintf("v1/games/icons?universeIds=%s&returnPolicy=%s&size=%s&format=%s&isCircular=%t",
			universeID, returnPolicy, size, format, isCircular), &tnr,
	)
	if err != nil {
		return Thumbnail{}, err
	}

	return tnr.Data[0], nil
}
