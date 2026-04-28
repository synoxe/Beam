package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"beam/internal/protocol"
)

func BuildMetadata(path string, checksum string) (*protocol.FileMetadata, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("verilen yol bir dosya değil, klasör")
	}
	return &protocol.FileMetadata{
		FileName: filepath.Base(path),
		Size:     info.Size(),
		Checksum: checksum,
	}, nil
}

func SanitizeFileName(name string) string {
	name = filepath.Base(name)

	replacer := strings.NewReplacer(
		"..", "",
		"/", "",
		"\\", "",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	name = replacer.Replace(name)
	name = strings.TrimSpace(name)

	if name == "" {
		return "received_file"
	}
	return name
}

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func UniqueFilePath(dir string, fileName string) string {
	fullPath := filepath.Join(dir, fileName)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fullPath
	}
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)

	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s_(%d)%s", base, i, ext)
		fullPath = filepath.Join(dir, candidate)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fullPath
		}
	}
}
