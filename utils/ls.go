package utils

import (
	"bergo/locales"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

type LsTool struct {
	Ig *Ignore
}

func (t *LsTool) List(path string) string {
	// 检查路径是否为文件
	if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
		if isHiddenFile(fileInfo.Name()) {
			return "" // 跳过隐藏文件
		}
		return fileInfo.Name()
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return locales.Sprintf("error reading directory: %v", err)
	}

	buf := bytes.NewBuffer(nil)
	fileBuf := bytes.NewBuffer(nil)
	dirBuf := bytes.NewBuffer(nil)
	for _, entry := range entries {
		name := entry.Name()

		// 跳过隐藏文件
		if isHiddenFile(name) {
			continue
		}

		if entry.IsDir() {
			name = name + "/" // 标记目录
		}

		if entry.IsDir() {
			dirBuf.WriteString(name)
			dirBuf.WriteString(" ")
		} else {
			fileBuf.WriteString(name)
			fileBuf.WriteString(" ")
		}
	}
	buf.WriteString(dirBuf.String())
	buf.WriteString(fileBuf.String())
	buf.WriteString("\n")
	return buf.String()
}

func (t *LsTool) ListWithPath(path string) []string {
	// 检查路径是否为文件
	if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
		if isHiddenFile(fileInfo.Name()) {
			return nil // 跳过隐藏文件
		}
		return []string{path}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	result := make([]string, 0)
	for _, entry := range entries {
		name := entry.Name()

		// 跳过隐藏文件
		if isHiddenFile(name) {
			continue
		}
		fullPath := filepath.Join(path, name)
		if entry.IsDir() {
			fullPath = fullPath + "/" // 标记目录
		}

		result = append(result, fullPath)
	}
	return result
}

func (t *LsTool) ListFilesRecursive(path string) string {
	// 检查路径是否为文件
	if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
		return path
	}

	return t.doWalk(path)
}

func (t *LsTool) doWalk(path string) string {
	if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
		return ""
	}
	buf := bytes.NewBuffer(nil)
	buf.WriteString(path)
	buf.WriteString(":\n")
	fileBuf := bytes.NewBuffer(nil)
	dirBuf := bytes.NewBuffer(nil)
	//遍历path
	entries, err := os.ReadDir(path)
	if err != nil {
		return ""
	}
	subDir := bytes.NewBuffer(nil)
	for _, entry := range entries {
		if isHiddenFile(entry.Name()) {
			continue
		}
		if t.Ig.MatchesPath(filepath.Join(path, entry.Name())) {
			continue
		}
		if entry.IsDir() {
			dirBuf.WriteString(entry.Name())
			dirBuf.WriteString("/ ")
			subDir.WriteString(t.doWalk(filepath.Join(path, entry.Name())))
			continue
		}
		fileBuf.WriteString(entry.Name())
		fileBuf.WriteString(" ")
	}
	buf.WriteString(dirBuf.String())
	buf.WriteString(fileBuf.String())
	buf.WriteString("\n")
	buf.Write(subDir.Bytes())
	return buf.String()
}

// isHiddenFile 检查文件名是否为隐藏文件
func isHiddenFile(name string) bool {
	return strings.HasPrefix(name, ".")
}
