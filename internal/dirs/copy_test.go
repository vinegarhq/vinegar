package dirs

import (
	"os"
	"testing"

	"path/filepath"
)

func createTestFile(t *testing.T, overlay, dest string) (string, string) {
	if err := Mkdirs(filepath.Join(overlay, "content"), filepath.Join(dest, "content")); err != nil {
		t.Fatal(err)
	}

	srcPath := filepath.Join(overlay, "content", "test_content.txt")
	destPath := filepath.Join(dest, "content", "test_content.txt")

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

	return srcPath, destPath
}

func test(t *testing.T, overlay, dest, srcPath, destPath string) {
	if err := OverlayDir(overlay, dest); err != nil {
		t.Fatal(err)
	}


	if res, err := CompareLink(srcPath, destPath); err != nil && !res {
		t.Fatal(err)
	}
}

func TestDefault(t *testing.T) {
	overlay := t.TempDir()
	dest := t.TempDir()

	srcPath, destPath := createTestFile(t, overlay, dest)
	test(t, overlay, dest, srcPath, destPath)
}

func TestAlreadyExists(t *testing.T) {
	overlay := t.TempDir()
	dest := t.TempDir()

	srcPath, destPath := createTestFile(t, overlay, dest)

	test(t, overlay, dest, srcPath, destPath)
	test(t, overlay, dest, srcPath, destPath)
}
