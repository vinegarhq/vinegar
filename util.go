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

type Dirs struct {
	Cache string
	Data  string
	Log   string
	Pfx   string
	Exe   string
}

func Errc(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func DeleteDir(dir ...string) {
	for _, d := range dir {
		log.Println("Deleting directory:", d)
		Errc(os.RemoveAll(d))
	}
}

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

func InitEnv(dirs *Dirs) {
	os.Setenv("WINEPREFIX", dirs.Pfx)
	os.Setenv("WINEARCH", "win64")
	// Removal of most unnecessary Wine facilities
	os.Setenv("WINEDEBUG", "fixme-all,-wininet,-ntlm,-winediag,-kerberos,+relay")
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

func Exec(dirs *Dirs, prog string, args ...string) {
	log.Println(args)
	cmd := exec.Command(prog, args...)
	timeFmt := time.Now().Format(time.RFC3339)

	stdoutFile, err := os.Create(filepath.Join(dirs.Log, timeFmt + "-stdout.log"))
	Errc(err)
	log.Println("Forwarding stdout to", stdoutFile.Name())
	defer stdoutFile.Close()

	stderrFile, err := os.Create(filepath.Join(dirs.Log, timeFmt + "-stderr.log"))
	Errc(err)
	log.Println("Forwarding stderr to", stderrFile.Name())
	defer stderrFile.Close()
	
	cmd.Dir = dirs.Cache
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	Errc(cmd.Run())

	logFile, err := stderrFile.Stat()
	Errc(err)
	if logFile.Size() == 0 {
		log.Println("Empty stderr log file detected, deleting")
		Errc(os.RemoveAll(stderrFile.Name()))
	}

	logFile, err = stdoutFile.Stat()
	Errc(err)
	if logFile.Size() == 0 {
		log.Println("Empty stdout file detected, deleted")
		Errc(os.RemoveAll(stdoutFile.Name()))
	}
}

func InitExec(dirs *Dirs, path string, url string, what string) (string) {
	path = filepath.Join(dirs.Exe, path)

	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		log.Println("Installing", what)
		Download(path, url)
	}

	if os.IsExist(err) {
		log.Println("Located executable:", path)
	}

	Errc(err)
	
	return path
}

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

	log.Println("Done downloading " + source)
}

func DownloadUnzip(source string, target string) {
	archive, err := zip.OpenReader(source)
	Errc(err)
	defer archive.Close()

	// Create blank file
	out, err := os.Create(target)
	Errc(err)
	defer out.Close()

	/* 
	 * We unzip only particularly a file, we dont create
	 * the zip structure.	
	 */
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
