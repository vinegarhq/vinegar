package main

import (
	"archive/zip"
	"compress/zlib"
	"encoding/gob"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sewnie/rbxbin"
	"github.com/sewnie/rbxweb"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
	"github.com/vinegarhq/vinegar/internal/netutil"
	"golang.org/x/sync/errgroup"

	. "github.com/pojntfx/go-gettext/pkg/i18n"
)

var (
	studio     = rbxweb.BinaryTypeWindowsStudio64
	channelKey = `HKCU\Software\ROBLOX Corporation\Environments\RobloxStudio\Channel`
)

type DeploymentPackage struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
}

type DeploymentData struct {
	Version  string                       `json:"version"`
	Packages map[string]DeploymentPackage `json:"packages"`
}

func (b *bootstrapper) updateDeployment() error {
	if err := b.setDeployment(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	slog.Info("Using Deployment",
		"guid", b.bin.GUID, "channel", b.bin.Channel)

	currentDeployment := DeploymentData{
		Version:  "",
		Packages: map[string]DeploymentPackage{},
	}

	deploymentPath := filepath.Join(dirs.Deployment, "deployment.vinegar")
	if _, err := os.Stat(deploymentPath); err == nil {
		err := func() error {
			file, err := os.Open(deploymentPath)
			if err != nil {
				return err
			}
			defer file.Close()

			decompressed, err := zlib.NewReader(file)
			if err != nil {
				slog.Error("Failed to decompress the Vinegar deployment file", "err", err)
				return nil
			}
			defer decompressed.Close()

			decoder := gob.NewDecoder(decompressed)
			err = decoder.Decode(&currentDeployment)
			if err != nil {
				slog.Error("Failed to decode the Vinegar deployment file", "err", err)
			}

			return nil
		}()

		if err != nil {
			return err
		}

		if currentDeployment.Version == b.bin.GUID {
			b.message(L("Up to date"), "guid", b.bin.GUID)
			return nil
		}
	}

	// No packages imply an inexisting or corrupt installation,
	// just to be safe delete everything and start over
	if len(currentDeployment.Packages) == 0 {
		slog.Warn("No packages installed, resetting deployment")
		if err := os.RemoveAll(dirs.Deployment); err != nil {
			return err
		}
	}

	b.message(L("Installing Studio"), "new", b.bin.GUID)

	if err := b.installDeployment(currentDeployment); err != nil {
		return err
	}

	defer b.performing()()

	b.message(L("Writing AppSettings"))
	if err := rbxbin.WriteAppSettings(dirs.Deployment); err != nil {
		return fmt.Errorf("appsettings: %w", err)
	}

	// Required for Studio to recognize its own channel:
	// https://github.com/vinegarhq/vinegar/issues/649
	//
	// Default channel is none, but UserChannel will set LIVE.
	if b.bin.Channel != "" && b.bin.Channel != "LIVE" {
		b.message(L("Writing Registry"))
		if err := b.pfx.RegistryAdd(channelKey, "www.roblox.com", b.bin.Channel); err != nil {
			return fmt.Errorf("set channel reg: %w", err)
		}
	}

	slog.Info("Successfully installed!", "guid", b.bin.GUID)
	return nil
}

func (b *bootstrapper) setDeployment() error {
	defer b.performing()()

	if b.cfg.Studio.ForcedVersion != "" {
		b.bin = &rbxbin.Deployment{
			Type:    studio,
			Channel: b.cfg.Studio.Channel,
			GUID:    b.cfg.Studio.ForcedVersion,
		}
		return nil
	}

	b.message(L("Checking for updates"))

	d, err := rbxbin.GetDeployment(b.rbx, studio, b.cfg.Studio.Channel)
	if err != nil {
		return err
	}

	gutil.IdleAdd(func() {
		b.info.SetLabel(d.Channel)
	})

	b.bin = d
	return nil
}

func (b *bootstrapper) installDeployment(previousDeployment DeploymentData) error {
	stop := b.performing()

	b.message(L("Finding Mirror"))
	m, err := rbxbin.GetMirror()
	if err != nil {
		return fmt.Errorf("fetch mirror: %w", err)
	}

	b.message(L("Fetching Package List"))
	pkgs, err := m.GetPackages(b.bin)
	if err != nil {
		return fmt.Errorf("fetch packages: %w", err)
	}

	sums := make([]string, len(pkgs))
	for _, pkg := range pkgs {
		sums = append(sums, pkg.Checksum)
	}

	// Prioritize smaller files first, to have less pressure
	// on network and extraction
	//
	// *Theoretically*, this should be better
	sort.SliceStable(pkgs, func(i, j int) bool {
		return pkgs[i].ZipSize < pkgs[j].ZipSize
	})

	// Create a temporary directory to store package files and
	// delete it when we are done installing
	downloadPath, err := os.MkdirTemp("", "vinegar-downloads")
	if err != nil {
		return err
	}
	defer os.RemoveAll(downloadPath)

	newDeployment := DeploymentData{
		Version:  b.bin.GUID,
		Packages: map[string]DeploymentPackage{},
	}

	newPkgs := []rbxbin.Package{}

	// Packages that exist in previous deployment are already
	// installed, so we just copy their data into the new
	// deployment's metadata
	for _, pkg := range pkgs {
		if installedPkg, ok := previousDeployment.Packages[pkg.Checksum]; !ok {
			newPkgs = append(newPkgs, pkg)
		} else {
			newDeployment.Packages[pkg.Checksum] = installedPkg
		}
	}

	b.message(L("Fetching Installation Directives"))
	pd, err := m.BinaryDirectories(b.bin)
	if err != nil {
		return fmt.Errorf("fetch package dirs: %w", err)
	}

	stop()

	// Uninstall packages that no longer exist in the newer version,
	// this uses the file list we generate on installation
	stalePkgs := []DeploymentPackage{}
	for checksum, pkgData := range previousDeployment.Packages {
		if _, ok := newDeployment.Packages[checksum]; !ok {
			stalePkgs = append(stalePkgs, pkgData)
		}
	}

	total := len(stalePkgs)
	if total > 0 {
		finished := int64(0)
		group := new(errgroup.Group)

		b.message(L("Cleaning Packages"))
		for _, pkg := range stalePkgs {
			group.Go(func() error {
				slog.Info("Uninstalling package", "name", pkg.Name)
				for _, file := range pkg.Files {
					path := filepath.Join(dirs.Deployment, file)

					_, err := os.Stat(path)
					if err == nil {
						err := os.Remove(path)
						if err != nil {
							return err
						}
					}
				}

				atomic.AddInt64(&finished, 1)
				gutil.IdleAdd(func() {
					b.pbar.SetFraction(float64(finished) / float64(total))
				})

				return nil
			})
		}

		if err := group.Wait(); err != nil {
			return err
		}
	}

	return b.installPackages(&m, newPkgs, pd, downloadPath, newDeployment)
}

func (b *bootstrapper) installPackages(
	mirror *rbxbin.Mirror,
	pkgs []rbxbin.Package,
	pdirs rbxbin.PackageDirectories,
	downloadPath string,
	newDeployment DeploymentData,
) error {
	total := len(pkgs)
	finished := int64(0)
	group := new(errgroup.Group)

	b.message(L("Installing Packages"), "count", len(pkgs), "dir", dirs.Deployment)
	deploymentMutex := sync.Mutex{}
	for _, pkg := range pkgs {
		group.Go(func() error {
			if err := b.installPackage(mirror, pdirs, &pkg, downloadPath, &newDeployment, &deploymentMutex); err != nil {
				return err
			}

			atomic.AddInt64(&finished, 1)
			gutil.IdleAdd(func() {
				b.pbar.SetFraction(float64(finished) / float64(total))
			})

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(dirs.Deployment, "deployment.vinegar"))
	if err != nil {
		return err
	}
	defer file.Close()

	compressed := zlib.NewWriter(file)
	defer compressed.Close()

	encoder := gob.NewEncoder(compressed)
	err = encoder.Encode(newDeployment)
	if err != nil {
		return err
	}

	return nil
}

func (b *bootstrapper) installPackage(
	mirror *rbxbin.Mirror,
	pdirs rbxbin.PackageDirectories,
	pkg *rbxbin.Package,
	downloadPath string,
	newDeployment *DeploymentData,
	deploymentMutex *sync.Mutex,
) error {
	src := filepath.Join(downloadPath, pkg.Checksum)
	dst, ok := pdirs[pkg.Name]
	if !ok {
		return fmt.Errorf("unhandled: %s", pkg.Name)
	}

	if err := pkg.Verify(src); err != nil {
		url := mirror.PackageURL(b.bin, pkg.Name)
		slog.Info("Downloading package", "name", pkg.Name, "sum", pkg.Checksum)
		if err := netutil.Download(url, src); err != nil {
			return err
		}
		if err := pkg.Verify(src); err != nil {
			return err
		}
	}

	// List all files in the package and store their paths
	// in the metadata file
	err := func() error {
		r, err := zip.OpenReader(src)
		if err != nil {
			return err
		}
		defer r.Close()

		filePaths := []string{}
		for _, f := range r.File {
			if !f.FileInfo().IsDir() {
				filePaths = append(filePaths, filepath.Join(dst, strings.ReplaceAll(f.Name, `\`, "/")))
			}
		}

		deploymentMutex.Lock()
		newDeployment.Packages[pkg.Checksum] = DeploymentPackage{
			Name:  pkg.Name,
			Files: filePaths,
		}
		deploymentMutex.Unlock()

		return nil
	}()

	if err != nil {
		return err
	}

	slog.Info("Extracting package", "name", pkg.Name, "dest", dst)
	return pkg.Extract(src, filepath.Join(dirs.Deployment, dst))
}
