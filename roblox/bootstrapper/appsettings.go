package bootstrapper

import (
	"log"
	"os"
	"path/filepath"
)

// WriteAppSettings writes the AppSettings.xml file - required
// to run Roblox - to the named directory, which is assumed to be
// a Binary's deployment directory.
func WriteAppSettings(dir string) error {
	log.Printf("Writing AppSettings: %s", dir)

	f, err := os.Create(filepath.Join(dir, "AppSettings.xml"))
	if err != nil {
		return err
	}
	defer f.Close()

	appSettings := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n" +
		"<Settings>\r\n" +
		"        <ContentFolder>content</ContentFolder>\r\n" +
		"        <BaseUrl>http://www.roblox.com</BaseUrl>\r\n" +
		"</Settings>\r\n"

	if _, err := f.WriteString(appSettings); err != nil {
		return err
	}

	return nil
}
