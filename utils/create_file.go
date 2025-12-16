package utils

import (
	"os"
	"path/filepath"
)

type CreateFile struct {
}

func (c *CreateFile) Create(path string) error {
	// 获取文件存在的目录
	dir := filepath.Dir(path)
	// 判断目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 创建目录
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func (c *CreateFile) CreateIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.Create(path)
	}
	return nil
}
