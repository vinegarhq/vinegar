package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func VerifyFileMD5(name string, sum string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	fsum := hex.EncodeToString(h.Sum(nil))

	if sum != fsum {
		return fmt.Errorf("file %s checksum mismatch: %s != %s", name, sum, fsum)
	}

	return nil
}
