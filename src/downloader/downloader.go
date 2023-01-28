//This file is for securely downloading rbxfpsunlocker and the launcher
//when not in a flatpak environment.

package downloader

import ("net/http"
	"io"
	"os"
	"log"
	"sync"
	"path/filepath"
	"archive/zip"
	"strings"
)

// Download from SOURCE (https) to TARGET
// Note: this is potentially DESTRUCTIVE. Be cautious where the target is.
// Sanity check is to be done by the caller.
func Download(source string, target string){
	target, err := filepath.Abs(target)
	if err != nil {
		log.Fatal("Failed to resolve download target. BUG!")
	}

	out, err := os.Create(target)
	if err != nil {
		//add zenity warning here
		log.Fatal("Failed to create required files, cannot continue with installation!")
	}
	defer out.Close()
	resp, err := http.Get(source)
	if (err != nil) && (resp.StatusCode != http.StatusOK){
		// Give up, since other code will break if this fails.
		//add zenity warning here
		log.Fatal("Failed to download required files, cannot continue with installation!")
	}
	defer resp.Body.Close()
	
	if _,err := io.Copy(out, resp.Body); err != nil {
		//add zenity warning here
		log.Fatal("Failed to write required files, cannot continue with installation!")
	}
	log.Println("Done downloading " + source)
}

//adapted from https://www.geeksforgeeks.org/how-to-uncompress-a-file-in-golang/
func Unzip(srcfile string, targetdir string)(error) {
	var filenames []string
	reader, err := zip.OpenReader(srcfile)
	if err != nil {
		log.Fatal("Problem opening zip. Something is wrong upstream!")
	}
	defer reader.Close()
	for _, file := range r.File{
		filepath := filepath.Join(destination, file.Name)
		if !strings.HasPrefix(filepath, filepath.Clean(destination) + string(os.PathSeparator)) {
	