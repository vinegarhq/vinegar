package splash

type Config struct {
	Enabled     bool   `toml:"enabled"`     // Determines if splash is shown or not
	LogoPath    string `toml:"logo_path"`   // Logo file path used to load and render the logo
	Style       string `toml:"style"`       // Style to use for the splash layout
	BgColor     uint32 `toml:"background"`  // Foreground color
	FgColor     uint32 `toml:"foreground"`  // Background color
	CancelColor uint32 `toml:"cancel,red"`  // Background color for the Cancel button
	AccentColor uint32 `toml:"accent"`      // Color for progress bar's track and ShowLog button
	TrackColor  uint32 `toml:"track,gray1"` // Color for the progress bar's background
	InfoColor   uint32 `toml:"info,gray2"`  // Foreground color for the text containing binary information
}
