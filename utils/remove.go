package utils

import (
	"os"
	"path/filepath"
	"strings"
)

type Remove struct {
	Root string
}

func (r *Remove) OutSideRoot(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return true
	}

	absRoot, err := filepath.Abs(r.Root)
	if err != nil {
		return true
	}

	// Check if the path is within the Root directory
	return !strings.HasPrefix(absPath, absRoot)
}
func (r *Remove) Do(path string) error {
	return os.Remove(path)
}
