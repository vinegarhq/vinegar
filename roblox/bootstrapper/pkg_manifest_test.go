package bootstrapper

import (
	"errors"
	"strconv"
	"testing"

	"github.com/vinegarhq/vinegar/roblox/version"
)

func TestParsePackages(t *testing.T) {
	manifest := []string{
		"v0",

		"foo.zip",
		"026b271a21b03f2e564c036525356db5",
		"71367142",
		"109436874",

		"bar.zip",
		"4d9ec7b52a29c80f3ce1f6a65b14b563",
		"408629",
		"1191394",
	}

	pkgs, err := parsePackages(manifest)
	if err != nil {
		t.Fatal(err)
	}

	pkgFooWant := Package{
		Name:     "foo.zip",
		Checksum: "026b271a21b03f2e564c036525356db5",
		Size:     109436874,
		ZipSize:  71367142,
	}

	pkgBarWant := Package{
		Name:     "bar.zip",
		Checksum: "4d9ec7b52a29c80f3ce1f6a65b14b563",
		Size:     1191394,
		ZipSize:  408629,
	}

	if pkgs[0] != pkgFooWant {
		t.Fatalf("package %v, want package match for %v", pkgs[0], pkgFooWant)
	}

	if pkgs[1] != pkgBarWant {
		t.Fatalf("package %v, want package match for %v", pkgs[0], pkgBarWant)
	}
}

func TestInvalidPackagePackageManifest(t *testing.T) {
	manifest := []string{
		"v0",
		"foo.zip",
		"026b271a21b03f2e564c036525356db5",
		"71367142",
	}

	_, err := parsePackages(manifest)
	if !errors.Is(err, ErrInvalidPkgManifest) {
		t.Fail()
	}

	manifest = append(manifest, "foo")

	_, err = parsePackages(manifest)
	if !errors.Is(err, strconv.ErrSyntax) {
		t.Fail()
	}
}

func TestUnhandledPackagePackageManifest(t *testing.T) {
	manifest := []string{
		"v1",
		"foo.zip",
		"026b271a21b03f2e564c036525356db5",
		"71367142",
		"109436874",
	}

	_, err := parsePackages(manifest)
	if !errors.Is(err, ErrUnhandledPkgManifestVer) {
		t.Fail()
	}
}

func TestExcludedPackage(t *testing.T) {
	manifest := []string{
		"v0",

		"WebView2RuntimeInstaller.zip",
		"e42a6697bf05466d4dba26c8fe476d2e",
		"1486447",
		"1589080",

		"RobloxPlayerLauncher.exe",
		"bcfb5b5e9e780e7ef4d281eb0efed185",
		"4974576",
		"4974576",
	}

	pkgs, err := parsePackages(manifest)
	if err != nil {
		t.Fatal(err)
	}

	if len(pkgs) != 0 {
		t.Fail()
	}
}

func TestChannelPath(t *testing.T) {
	if channelPath(version.DefaultChannel) != "/" {
		t.Fatal("expected default channel empty path")
	}

	if channelPath("ZLive") != "/channel/zlive/" {
		t.Fatal("expected channel path")
	}
}
