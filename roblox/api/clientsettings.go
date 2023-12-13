package api

// ClientVersion is a representation of the Roblox ClientVersionResponse model.
type ClientVersion struct {
	Version                 string `json:"version"`
	ClientVersionUpload     string `json:"clientVersionUpload"`
	BootstrapperVersion     string `json:"bootstrapperVersion"`
	NextClientVersionUpload string `json:"nextClientVersionUpload,omitempty"`
	NextClientVersion       string `json:"nextClientVersion,omitempty"`
}

// GetClientVersion gets the ClientVersion for the named binaryType and deployment channel.
func GetClientVersion(binaryType string, channel string) (ClientVersion, error) {
	var cv ClientVersion

	ep := "v2/client-version/" + binaryType
	if channel != "" {
		ep += "/channel/" + channel
	}

	err := Request("GET", "clientsettings", ep, &cv)
	if err != nil {
		return ClientVersion{}, err
	}

	return cv, nil
}
