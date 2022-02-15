package fzd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	HeadFileName = "HEAD"
)

func writeHead(basePath string, name string) error {
	path := filepath.Join(basePath, HeadFileName)
	err := os.WriteFile(path, []byte(name), 0600)
	if err != nil {
		return fmt.Errorf("failed to write %v: %w", path, err)
	}
	return nil
}

func readHead(basePath string) (string, error) {
	path := filepath.Join(basePath, HeadFileName)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", ErrIndexHeadDoesNotExist
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %v: %w", path, err)
	}
	name := string(content)
	return name, nil
}
