// Copyright vinegar-development 2023

package vinegar

import (
	"log"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"net/http"
	"archive/zip"
)

// Primary struct keeping track of vinegar's directories.
type Dirs struct {
	Cache string
	Data  string
	Log   string
	Pfx   string
	Exe   string
}

// Helper function to handle error failure
func Errc(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Deletes directories, with logging(!)
func DeleteDir(dir ...string) {
	for _, d := range dir {
		log.Println("Deleting directory:", d)
		Errc(os.RemoveAll(d))
	}
}

// Check for directories if they exist, if not, create them with 0755,
// and let the user know with logging.
func DirsCheck(dir ...string) {
	for _, d := range dir {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			log.Println("Creating directory:", d)
		} else { 
			continue
		}
		Errc(os.MkdirAll(d, 0755))
	}
}

// Creates the Dirs struct with its initial values.
func InitDirs() *Dirs {
	dirs := new(Dirs)
	home := os.Getenv("HOME")

	if home == "" {
		log.Fatal("Failed to get $HOME variable")
	}

	dirs.Cache = (filepath.Join(home, ".cache", "/vinegar"))
	dirs.Data  = (filepath.Join(home, ".local", "share", "/vinegar"))
	dirs.Log   = (filepath.Join(dirs.Cache, "/logs"))
	dirs.Pfx   = (filepath.Join(dirs.Data, "/pfx"))
	dirs.Exe   = (filepath.Join(dirs.Cache, "/exe"))

	return dirs
}

// Main environment initialization for Wine, DXVK, and PRIME.
func InitEnv(dirs *Dirs) {
	os.Setenv("WINEPREFIX", dirs.Pfx)
	os.Setenv("WINEARCH", "win64") // Required for rbxfpsunlocker
	// Removal of most unnecessary Wine facilities
	os.Setenv("WINEDEBUG", "fixme-all,-wininet,-ntlm,-winediag,-kerberos")
	os.Setenv("WINEDLLOVERRIDES", "dxdiagn=d;winemenubuilder.exe=d")
	os.Setenv("DXVK_LOG_LEVEL", "warn")
	os.Setenv("DXVK_LOG_PATH", "none")
	os.Setenv("DXVK_STATE_CACHE_PATH", filepath.Join(dirs.Cache, "dxvk"))
	
	os.Setenv("MESA_GL_VERSION_OVERRIDE", "4.4")
	// Use the dedicated gpu if available, untested
	os.Setenv("DRI_PRIME", "1")
	os.Setenv("__NV_PRIME_RENDER_OFFLOAD", "1")
	os.Setenv("__VK_LAYER_NV_optimus", "NVIDIA_only")
	os.Setenv("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
}

// Execute a program with arguments whilst keeping 
// it's stderr output to a log file, stdout is ignored and is sent to os.Stdout.
func Exec(dirs *Dirs, prog string, args ...string) {
	log.Println(args)
	cmd := exec.Command(prog, args...)
	timeFmt := time.Now().Format(time.RFC3339)

	stderrFile, err := os.Create(filepath.Join(dirs.Log, timeFmt + "-stderr.log"))
	Errc(err)
	log.Println("Forwarding stderr to", stderrFile.Name())
	defer stderrFile.Close()
	
	cmd.Dir = dirs.Cache

	// Stdout is particularly always empty, so we forward its to ours
	cmd.Stdout = os.Stdout
	cmd.Stderr = stderrFile

	Errc(cmd.Run())

	logFile, err := stderrFile.Stat()
	Errc(err)
	if logFile.Size() == 0 {
		log.Println("Empty stderr log file detected, deleting")
		Errc(os.RemoveAll(stderrFile.Name()))
	}
}

// Download a single file into a target file.
func Download(source string, target string){
	// Create blank file
	out, err := os.Create(target)
	Errc(err)
	defer out.Close()

	resp, err := http.Get(source)
	Errc(err)
	defer resp.Body.Close()
	
	_, err = io.Copy(out, resp.Body)
	Errc(err)

	log.Println("Downloaded", source)
}

// Unzip a single file without keeping track of zip's structue into 
// a target file, Will remove the source zip file after successful extraction.
func Unzip(source string, target string) {
	archive, err := zip.OpenReader(source)
	Errc(err)
	defer archive.Close()

	// Create blank file
	out, err := os.Create(target)
	Errc(err)
	defer out.Close()

	for _, zipped := range archive.File {
		zippedFile, err := zipped.Open()
		Errc(err)
		defer zippedFile.Close()

		_, err = io.Copy(out, zippedFile)
		Errc(err)

		log.Println("Unzipped", zipped.Name)
	}

	Errc(os.RemoveAll(source))
	log.Println("Removed archive", source)
}
