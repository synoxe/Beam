package hashutil

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha256.New()

	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}
	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum), nil
}
