//This file is for securely downloading rbxfpsunlocker and the launcher
//when not in a flatpak environment.

package downloader

import ("net/http"
	"io"
	"os"
)

// Download from SOURCE (https) to TARGET
// Note: this is potentially DESTRUCTIVE. Be cautious where the target is.
func Download(source string, target string){
	out, err := os.Create(target)
	if err != nil {
		panic("Failed to create required files, cannot continue with installation!")
	}
	defer out.Close()
	resp, err := http.Get(source)
	if (err != nil) && (resp.StatusCode != http.StatusOK){
		// Give up, since other code will break if this fails.
		panic("Failed to download required files, cannot continue with installation!")
	}
	defer resp.Body.Close()
	
	if _,err := io.Copy(out, resp.Body); err != nil {
		panic("Failed to write required files, cannot continue with installation!")
	}

}