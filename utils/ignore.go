package utils

import (
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"
)

type Ignore struct {
	ignores []*ignore.GitIgnore
}

func NewIgnore(rootPath string, ignoreFiles []string) *Ignore {
	ignores := make([]*ignore.GitIgnore, 0)
	for _, ignoreFile := range ignoreFiles {
		ignorePath := filepath.Join(rootPath, ignoreFile)
		gi, err := ignore.CompileIgnoreFile(ignorePath)
		if err == nil {
			ignores = append(ignores, gi)
		}
	}
	return &Ignore{ignores: ignores}
}

func (ig *Ignore) MatchesPath(path string) bool {
	// Convert to slash-separated path for consistency
	path = filepath.ToSlash(path)
	for _, gi := range ig.ignores {
		if gi.MatchesPath(path) {
			return true
		}
	}
	return false
}
