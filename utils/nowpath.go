package utils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type NowPath struct {
	current string
	t       *Trie
}

func NewNowPath() *NowPath {
	return &NowPath{
		current: "",
		t:       NewTrie(),
	}
}

func (n *NowPath) Update(path string) {
	dir := filepath.Dir(path)
	if dir == n.current {
		return
	}
	if _, err := os.Stat(dir); err != nil {
		n.current = ""
		n.t = NewTrie()
		return
	}
	n.current = dir
	n.t = NewTrie()
	paths, _ := os.ReadDir(dir)
	for _, path := range paths {
		if strings.HasPrefix(path.Name(), ".") {
			continue
		}
		if path.IsDir() {
			n.t.Put(path.Name()+"/", filepath.Join(dir, path.Name())) // 目录结尾加/
		} else {
			n.t.Put(path.Name(), filepath.Join(dir, path.Name()))
		}
	}
}

type MatchFile struct {
	Path string
	Name string
}

func (n *NowPath) MatchFiles(prefix string) []*MatchFile {
	if n.t == nil {
		return nil
	}
	base := filepath.Base(prefix)
	if strings.HasSuffix(prefix, "/") || base == "." || base == ".." {
		base = ""
	}
	var files []*MatchFile
	var dirs []*MatchFile
	n.t.WalkPath(base, func(key string, value interface{}) {
		if strings.HasSuffix(key, "/") {
			dirs = append(dirs, &MatchFile{
				Path: value.(string) + "/",
				Name: key,
			})
		} else {
			files = append(files, &MatchFile{
				Path: value.(string),
				Name: key,
			})
		}
	})
	// sort用以保持目录和文件的顺序
	sort.SliceStable(dirs, func(i, j int) bool {
		return dirs[i].Name < dirs[j].Name
	})
	sort.SliceStable(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})
	return append(dirs, files...)
}
