package api

// Creator is a representation of the Roblox GameCreator model.
type Creator struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	IsRNVAccount     bool   `json:"isRNVAccount"`
	HasVerifiedBadge bool   `json:"hasVerifiedBadge"`
}

// GameDetail is a representation of the Roblox GameDetailResponse model.
type GameDetail struct {
	ID                        int64    `json:"id"`
	RootPlaceID               int64    `json:"rootPlaceId"`
	Name                      string   `json:"name"`
	Description               string   `json:"description"`
	SourceName                string   `json:"sourceName"`
	SourceDescription         string   `json:"sourceDescription"`
	Creator                   Creator  `json:"creator"`
	Price                     int64    `json:"price"`
	AllowedGearGenres         []string `json:"allowedGearGenres"`
	AllowedGearCategories     []string `json:"allowedGearCategories"`
	IsGenreEnforced           bool     `json:"isGenreEnforced"`
	CopyingAllowed            bool     `json:"copyingAllowed"`
	Playing                   int64    `json:"playing"`
	Visits                    int64    `json:"visits"`
	MaxPlayers                int32    `json:"maxPlayers"`
	Created                   string   `json:"created"`
	Updated                   string   `json:"updated"`
	StudioAccessToApisAllowed bool     `json:"studioAccessToApisAllowed"`
	CreateVipServersAllowed   bool     `json:"createVipServersAllowed"`
	UniverseAvatarType        string   `json:"universeAvatarType"`
	Genre                     string   `json:"genre"`
	IsAllGenre                bool     `json:"isAllGenre"`
	IsFavoritedByUser         bool     `json:"isFavoritedByUser"`
	FavoritedCount            int64    `json:"favoritedCount"`
}

// GameDetail is a representation of the Roblox ApiArrayResponse GameDetailResponse model.
type GameDetailsResponse struct {
	Data []GameDetail `json:"data"`
}

// GetGameDetails gets the list of a named universeID's detail
func GetGameDetails(universeID string) (GameDetail, error) {
	var gdr GameDetailsResponse

	// uids := strings.Join(universeIDs, ",")
	err := Request("GET", "games", "v1/games?universeIds="+universeID, &gdr)
	if err != nil {
		return GameDetail{}, err
	}

	return gdr.Data[0], nil
}
