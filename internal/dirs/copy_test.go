package dirs

import (
	"os"
	"testing"

	"path/filepath"
)

func TestDefault(t *testing.T) {
	overlay := t.TempDir()
	dest := t.TempDir()

	if err := Mkdirs(filepath.Join(overlay, "content"), filepath.Join(dest, "content")); err != nil {
		t.Fatal(err)
	}

	srcPath := filepath.Join(overlay, "content", "test_content.txt")
	file, err :=  os.OpenFile(srcPath, os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.WriteString("MEOWMEOW TESTING YAHH!!"); err != nil {
		file.Close()
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	err = OverlayDir(overlay, dest)

	if err != nil {
		t.Fatal(err)
	}

	destPath := filepath.Join(dest, "content", "test_content.txt")

	if res, err := CompareLink(srcPath, destPath); err != nil && !res {
		t.Fatal(err)
	}
}

func TestAlreadyExists(t *testing.T) {
	overlay := t.TempDir()
	dest := t.TempDir()

	if err := Mkdirs(filepath.Join(overlay, "content"), filepath.Join(dest, "content")); err != nil {
		t.Fatal(err)
	}

	srcPath := filepath.Join(overlay, "content", "test_content.txt")
	file, err :=  os.OpenFile(srcPath, os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.WriteString("MEOWMEOW TESTING YAHH!!"); err != nil {
		file.Close()
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	err = OverlayDir(overlay, dest)

	if err != nil {
		t.Fatal(err)
	}

	destPath := filepath.Join(dest, "content", "test_content.txt")

	if res, err := CompareLink(srcPath, destPath); err != nil && !res {
		t.Fatal(err)
	}

	err = OverlayDir(overlay, dest)

	if err != nil {
		t.Fatal(err)
	}

	if res, err := CompareLink(srcPath, destPath); err != nil && !res {
		t.Fatal(err)
	}
}
