package utils

import (
	"bergo/locales"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Checkpoint struct {
	workspacePath  string
	shadowRepoPath string
}

func NewCheckpoint(workPath string, shadowRepoPath string) *Checkpoint {
	return &Checkpoint{
		workspacePath:  workPath,
		shadowRepoPath: shadowRepoPath,
	}
}

// 使用git初始化一个影子仓库，并且使用git config把core.worktree指向workspacePath
func (c *Checkpoint) InitShadowRepo() error {
	// 检查workspacePath是否存在
	if _, err := os.Stat(c.workspacePath); err != nil {
		return locales.Errorf("workspace path %s does not exist: %v", c.workspacePath, err)
	}
	// 检查shadowRepoPath是否存在
	if _, err := os.Stat(c.shadowRepoPath); err == nil {
		return nil
	}

	// 创建影子仓库目录
	if err := os.MkdirAll(c.shadowRepoPath, 0755); err != nil {
		return locales.Errorf("fail to create shadow repo: %w", err)
	}

	// 初始化git仓库
	cmd := exec.Command("git", "init")
	cmd.Dir = c.shadowRepoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return locales.Errorf("fail to init shadow repo: %w, stderr output: %s", err, string(output))
	}

	// 设置worktree
	cmd = exec.Command("git", "config", "core.worktree", c.workspacePath)
	cmd.Dir = c.shadowRepoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return locales.Errorf("fail to set worktree: %w, stderr output: %s", err, string(output))
	}
	return nil
}

// 保存快照并返回commit hash
func (c *Checkpoint) Save(comment string) (string, error) {
	shadowRepo := c.shadowRepoPath
	c.changedotgitFileName()
	defer c.revertotgitFileName()

	// 添加所有变更
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = shadowRepo
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", locales.Errorf("fail to add all changes: %w, stderr output: %s", err, string(output))
	}

	// 提交变更
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", comment)
	cmd.Dir = shadowRepo
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", locales.Errorf("fail to commit changes: %w, stderr output: %s", err, string(output))
	}

	// 获取commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = shadowRepo
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", locales.Errorf("fail to get commit hash: %w, stderr output: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// 回退到指定commit
func (c *Checkpoint) Revert(hash string) error {
	shadowRepo := c.shadowRepoPath
	c.changedotgitFileName()
	defer c.revertotgitFileName()

	cmd := exec.Command("git", "reset", "--hard", hash)
	cmd.Dir = shadowRepo
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		return locales.Errorf("fail to revert to commit: %w, stderr output: %s", err, string(output))
	}

	cmd = exec.Command("git", "clean", "-fd")
	cmd.Dir = shadowRepo
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		return locales.Errorf("fail to clean  commit: %w, stderr output: %s", err, string(output))
	}
	return nil
}

// 重命名嵌套.git目录
func (c *Checkpoint) changedotgitFileName() {
	filepath.Walk(c.workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		relPath, err := filepath.Rel(c.workspacePath, path)
		if err != nil {
			return nil
		}
		if relPath == ".git" {
			return nil
		}
		if info.IsDir() && info.Name() == ".git" {
			newPath := filepath.Join(filepath.Dir(path), ".git_bergo_disabled")
			os.Rename(path, newPath)
		}
		return nil
	})
}

// 恢复.git目录名
func (c *Checkpoint) revertotgitFileName() {
	filepath.Walk(c.workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".git_bergo_disabled" {
			newPath := filepath.Join(filepath.Dir(path), ".git")
			os.Rename(path, newPath)
		}
		return nil
	})
}

// 判断工作区是否有未提交的变更
func (c *Checkpoint) HasChange() (bool, error) {
	c.changedotgitFileName()
	defer c.revertotgitFileName()

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = c.shadowRepoPath
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, locales.Errorf("fail to get git status: %w, stderr output: %s", err, string(out))
	}
	return len(out) > 0, nil
}

// 创建一个分支
func (c *Checkpoint) NewBranch(branch string) error {
	shadowRepo := c.shadowRepoPath
	c.changedotgitFileName()
	defer c.revertotgitFileName()

	// 创建新分支
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = shadowRepo
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		return locales.Errorf("fail to create new branch: %w, stderr output: %s", err, string(output))
	}
	return nil
}

// 切换到指定分支
func (c *Checkpoint) Checkout(branch string) error {
	shadowRepo := c.shadowRepoPath
	c.changedotgitFileName()
	defer c.revertotgitFileName()

	// 切换到指定分支
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = shadowRepo
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		return locales.Errorf("fail to checkout branch: %w, stderr output: %s", err, string(output))
	}
	return nil
}

// 获取当前git分支
func (c *Checkpoint) Branch() (string, error) {
	shadowRepo := c.shadowRepoPath
	c.changedotgitFileName()
	defer c.revertotgitFileName()

	// 获取当前分支名
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = shadowRepo
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", locales.Errorf("fail to get current branch: %w, stderr output: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func GetBergoHomeSpace() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	bergoPath := filepath.Join(userHomeDir, ".bergo")
	if _, err := os.Stat(bergoPath); os.IsNotExist(err) {
		os.Mkdir(bergoPath, 0755)
	}
	return bergoPath
}

func GetWorkspaceStorePath() string {
	workspaceStorePath, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	sha256Hex := sha256.Sum256([]byte(workspaceStorePath))
	hexStr := hex.EncodeToString(sha256Hex[:])

	path := filepath.Join(GetBergoHomeSpace(), hexStr)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
	return path

}
