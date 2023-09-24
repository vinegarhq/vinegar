package bootstrapper

import (
	"log"
	"os"
	"path/filepath"
)

func WriteAppSettings(dir string) error {
	log.Printf("Writing AppSettings: %s", dir)

	file, err := os.Create(filepath.Join(dir, "AppSettings.xml"))
	if err != nil {
		return err
	}
	defer file.Close()

	appSettings := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n" +
		"<Settings>\r\n" +
		"        <ContentFolder>content</ContentFolder>\r\n" +
		"        <BaseUrl>http://www.roblox.com</BaseUrl>\r\n" +
		"</Settings>\r\n"

	if _, err := file.WriteString(appSettings); err != nil {
		return err
	}

	return nil
}
