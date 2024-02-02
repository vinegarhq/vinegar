package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/config/editor"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/roblox"
)

var (
	BinPrefix string
	Version   string
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [-config filepath] player|studio exec|run [args...]")
	fmt.Fprintln(os.Stderr, "usage: vinegar [-config filepath] player|studio kill|winetricks")
	fmt.Fprintln(os.Stderr, "       vinegar [-config filepath] sysinfo")
	fmt.Fprintln(os.Stderr, "       vinegar delete|edit|version")
	os.Exit(1)
}

func main() {
	configPath := flag.String("config", filepath.Join(dirs.Config, "config.toml"), "config.toml file which should be used")
	flag.Parse()

	cmd := flag.Arg(0)
	args := flag.Args()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)

	switch cmd {
	case "delete", "edit", "version":
		switch cmd {
		case "delete":
			log.Println("Deleting Wineprefixes and Roblox Binary deployments!")

			if err := os.RemoveAll(dirs.Prefixes); err != nil {
				log.Fatal(err)
			}
		case "edit":
			if err := editor.Edit(*configPath); err != nil {
				log.Fatal(err)
			}
		case "version":
			fmt.Println("Vinegar", Version)
		}
	case "player", "studio", "sysinfo":
		DeleteOldPrefix() // Remove after a few releases

		cfg, err := config.Load(*configPath)
		if err != nil {
			log.Fatal(err)
		}

		if cmd == "sysinfo" {
			PrintSysinfo(&cfg)
			os.Exit(0)
		}

		var bt roblox.BinaryType
		switch cmd {
		case "player":
			bt = roblox.Player
		case "studio":
			bt = roblox.Studio
		}

		b, err := NewBinary(bt, &cfg)
		if err != nil {
			log.Fatal(err)
		}

		switch flag.Arg(1) {
		case "exec":
			if len(args) < 2 {
				usage()
			}

			if err := b.Prefix.Wine(args[2], args[3:]...).Run(); err != nil {
				log.Fatal(err)
			}
		case "kill":
			b.Prefix.Kill()
		case "winetricks":
			if err := b.Prefix.Winetricks(); err != nil {
				log.Fatal(err)
			}
		case "run":
			err := b.Main(args[2:]...)
			if err == nil {
				log.Println("Goodbye")
				os.Exit(0)
			}

			if !cfg.Splash.Enabled || b.Splash.IsClosed() {
				log.Fatal(err)
			}

			log.Println(err)
			b.Splash.SetMessage("Oops!")
			b.Splash.Dialog(fmt.Sprintf(DialogFailure, err), false)
			os.Exit(1)
		default:
			usage()
		}
	default:
		usage()
	}
}

func DeleteOldPrefix() {
	// The old prefix does not exist or has issues, return.
	_, err := os.Stat(dirs.Prefix)
	if err != nil {
		return
	}

	log.Println("Deleting deprecated old Wineprefix!")
	if err := os.RemoveAll(dirs.Prefix); err != nil {
		log.Fatal(err)
	}
}

func LogFile(name string) (*os.File, error) {
	if err := dirs.Mkdirs(dirs.Logs); err != nil {
		return nil, err
	}

	// name-2006-01-02T15:04:05Z07:00.log
	path := filepath.Join(dirs.Logs, name+"-"+time.Now().Format(time.RFC3339)+".log")

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s log file: %w", name, err)
	}

	log.Printf("Logging to file: %s", path)

	return file, nil
}
