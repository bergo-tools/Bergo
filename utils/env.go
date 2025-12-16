package utils

import (
	"bergo/locales"
	"fmt"
	"os"
	"strings"
)

// 检测当前目录是否是git仓库
func IsGitRepo() bool {
	_, err := os.Stat(".git")
	return !os.IsNotExist(err)
}

// 检测当前目录是否存在.bergo目录
func IsBergoInit() bool {
	_, err := os.Stat(".bergo")
	return !os.IsNotExist(err)
}

// 创建.bergo目录
func CreateBergoDir() error {
	return os.Mkdir(".bergo", 0755)
}

// 检测当前目录的.gitignore中是否包含.bergo
func IsGitIgnoreHasBergo() bool {
	content, err := os.ReadFile(".gitignore")
	if err != nil {
		return false
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == ".bergo" {
			return true
		}
	}
	return false
}

// 将.bergo添加到.gitignore中
func AddBergoToGitIgnore() error {
	// 检查是否已经存在
	if IsGitIgnoreHasBergo() {
		return nil
	}

	// 读取现有内容
	content, err := os.ReadFile(".gitignore")
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// 添加.bergo到内容中
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		content = append(content, '\n')
	}
	content = append(content, []byte(".bergo\n")...)

	// 写回文件
	return os.WriteFile(".gitignore", content, 0644)
}

func EnvInit() {
	if !IsGitRepo() {
		fmt.Println(locales.Sprintf("You workspace is not in a git repository"))
	}
	if !IsGitIgnoreHasBergo() {
		if err := AddBergoToGitIgnore(); err != nil {
			fmt.Println(locales.Sprintf("fail to add .bergo to .gitignore"))
		}
	}
}
