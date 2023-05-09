package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func Verify(filePath string, signature string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	if signature != hex.EncodeToString(hash.Sum(nil)) {
		return fmt.Errorf("file %s checksum mismatch: %x", filePath, hash.Sum(nil))
	}

	return nil
}
