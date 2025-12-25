package skills

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed builtin/*
var builtinSkills embed.FS

// ExtractBuiltinSkills 将内置skills释放到用户主目录的.bergoskills下
func ExtractBuiltinSkills() error {
	targetDir, err := GetSkillsPath()
	if err != nil {
		return err
	}

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills dir: %w", err)
	}

	// 获取内置skills列表，先清空这些目录
	entries, err := builtinSkills.ReadDir("builtin")
	if err != nil {
		return fmt.Errorf("failed to read builtin skills: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			skillPath := filepath.Join(targetDir, entry.Name())
			// 删除已存在的内置skill目录，确保更新
			os.RemoveAll(skillPath)
		}
	}

	// 遍历嵌入的文件并释放
	err = fs.WalkDir(builtinSkills, "builtin", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录 "builtin"
		if path == "builtin" {
			return nil
		}

		// 计算相对路径（去掉 "builtin/" 前缀）
		relPath, err := filepath.Rel("builtin", path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// 读取嵌入的文件内容
		content, err := builtinSkills.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// 写入目标文件
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to extract builtin skills: %w", err)
	}

	return nil
}
