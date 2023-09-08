package main

import (
	"log"
	"path/filepath"

	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/util"
	"github.com/vinegarhq/vinegar/wine"
)

func InstallWebview2(pfx *wine.Prefix) error {
	log.Println("Downloading Microsoft Edge WebView2 Evergreen Bootstrapper")
	installer := filepath.Join(dirs.Downloads, "MicrosoftEdgeWebview2Setup.exe")

	// just in case
	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	err := util.Download("https://go.microsoft.com/fwlink/p/?LinkId=2124703", installer)
	if err != nil {
		return err
	}

	log.Println("Launching Evergreen Bootstrapper")
	return pfx.ExecWine(installer)
}
