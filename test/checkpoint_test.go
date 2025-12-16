package test

import (
	"os"
	"path/filepath"
	"testing"

	"bergo/utils"
)

func TestCheckpoint_NewCheckpoint(t *testing.T) {
	workspacePath := "/tmp/workspace"
	shadowRepoPath := "/tmp/shadow"

	cp := utils.NewCheckpoint(workspacePath, shadowRepoPath)

	if cp == nil {
		t.Fatal("NewCheckpoint should not return nil")
	}

	// 由于字段是私有的，我们无法直接访问验证
	// 但我们可以通过方法调用来间接测试
}

func TestCheckpoint_InitShadowRepo(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "workspace")
	shadowRepoPath := filepath.Join(tempDir, "shadow")

	// 创建工作区目录
	err := os.MkdirAll(workspacePath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	cp := utils.NewCheckpoint(workspacePath, shadowRepoPath)

	// 测试初始化影子仓库
	err = cp.InitShadowRepo()
	if err != nil {
		t.Fatalf("InitShadowRepo failed: %v", err)
	}

	// 检查影子仓库目录是否创建
	if _, err := os.Stat(shadowRepoPath); os.IsNotExist(err) {
		t.Fatal("Shadow repo directory should be created")
	}

	// 检查是否初始化了git仓库
	gitDir := filepath.Join(shadowRepoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatal("Git repository should be initialized")
	}
}

func TestCheckpoint_HasChange(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "workspace")
	shadowRepoPath := filepath.Join(tempDir, "shadow")

	// 创建工作区目录
	err := os.MkdirAll(workspacePath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	cp := utils.NewCheckpoint(workspacePath, shadowRepoPath)

	// 初始化影子仓库
	err = cp.InitShadowRepo()
	if err != nil {
		t.Fatalf("InitShadowRepo failed: %v", err)
	}

	// 创建一个测试文件
	testFile := filepath.Join(workspacePath, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 测试是否有变更
	hasChange, err := cp.HasChange()
	if err != nil {
		t.Fatalf("HasChange failed: %v", err)
	}

	if !hasChange {
		t.Error("Should detect uncommitted changes")
	}
}

func TestCheckpoint_SaveAndRevert(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "workspace")
	shadowRepoPath := filepath.Join(tempDir, "shadow")

	// 创建工作区目录
	err := os.MkdirAll(workspacePath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// 创建一个测试文件
	testFile := filepath.Join(workspacePath, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cp := utils.NewCheckpoint(workspacePath, shadowRepoPath)

	// 初始化影子仓库
	err = cp.InitShadowRepo()
	if err != nil {
		t.Fatalf("InitShadowRepo failed: %v", err)
	}

	// 保存快照
	commitHash, err := cp.Save("Initial commit")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if commitHash == "" {
		t.Error("Commit hash should not be empty")
	}

	// 修改文件内容
	err = os.WriteFile(testFile, []byte("modified content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 再次保存
	newCommitHash, err := cp.Save("Second commit")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if newCommitHash == commitHash {
		t.Error("New commit hash should be different")
	}

	// 测试回退到第一个提交
	err = cp.Revert(commitHash)
	if err != nil {
		t.Fatalf("Revert failed: %v", err)
	}

	// 检查文件内容是否回退
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "initial content" {
		t.Errorf("File content should be reverted. Expected 'initial content', got '%s'", string(content))
	}
}

func TestCheckpoint_BranchOperations(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "workspace")
	shadowRepoPath := filepath.Join(tempDir, "shadow")

	// 创建工作区目录
	err := os.MkdirAll(workspacePath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// 创建一个测试文件
	testFile := filepath.Join(workspacePath, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cp := utils.NewCheckpoint(workspacePath, shadowRepoPath)

	// 初始化影子仓库
	err = cp.InitShadowRepo()
	if err != nil {
		t.Fatalf("InitShadowRepo failed: %v", err)
	}

	// 保存初始提交
	_, err = cp.Save("Initial commit")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 测试获取当前分支
	currentBranch, err := cp.Branch()
	if err != nil {
		t.Fatalf("Branch() failed: %v", err)
	}

	if currentBranch != "master" && currentBranch != "main" {
		t.Errorf("Expected 'master' or 'main', got '%s'", currentBranch)
	}

	// 测试创建新分支
	err = cp.NewBranch("feature-branch")
	if err != nil {
		t.Fatalf("NewBranch failed: %v", err)
	}

	// 验证分支切换成功
	currentBranch, err = cp.Branch()
	if err != nil {
		t.Fatalf("Branch() failed after NewBranch: %v", err)
	}

	if currentBranch != "feature-branch" {
		t.Errorf("Expected 'feature-branch', got '%s'", currentBranch)
	}

	// 在新分支上修改文件
	err = os.WriteFile(testFile, []byte("feature content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 保存新分支的修改
	featureCommit, err := cp.Save("Feature commit")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 切换回主分支
	err = cp.Checkout("master")
	if err != nil {
		// 尝试切换回main分支（新的git默认分支名）
		err = cp.Checkout("main")
		if err != nil {
			t.Fatalf("Checkout failed: %v", err)
		}
	}

	// 验证文件内容回到初始状态
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "initial content" {
		t.Errorf("File content should be reverted to initial content. Expected 'initial content', got '%s'", string(content))
	}

	// 再次切换到feature分支
	err = cp.Checkout("feature-branch")
	if err != nil {
		t.Fatalf("Checkout to feature-branch failed: %v", err)
	}

	// 验证文件内容是feature分支的内容
	content, err = os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "feature content" {
		t.Errorf("File content should be feature content. Expected 'feature content', got '%s'", string(content))
	}

	// 验证feature提交哈希不为空
	if featureCommit == "" {
		t.Error("Feature commit hash should not be empty")
	}
}

func TestCheckpoint_NewBranchError(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "workspace")
	shadowRepoPath := filepath.Join(tempDir, "shadow")

	// 创建工作区目录
	err := os.MkdirAll(workspacePath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	cp := utils.NewCheckpoint(workspacePath, shadowRepoPath)

	// 尝试在未初始化的仓库中创建分支（应该失败）
	err = cp.NewBranch("test-branch")
	if err == nil {
		t.Error("NewBranch should fail on uninitialized repo")
	}
}

func TestCheckpoint_CheckoutError(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "workspace")
	shadowRepoPath := filepath.Join(tempDir, "shadow")

	// 创建工作区目录
	err := os.MkdirAll(workspacePath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	cp := utils.NewCheckpoint(workspacePath, shadowRepoPath)

	// 尝试在未初始化的仓库中切换分支（应该失败）
	err = cp.Checkout("non-existent-branch")
	if err == nil {
		t.Error("Checkout should fail on uninitialized repo")
	}
}

func TestCheckpoint_BranchError(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "workspace")
	shadowRepoPath := filepath.Join(tempDir, "shadow")

	// 创建工作区目录
	err := os.MkdirAll(workspacePath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	cp := utils.NewCheckpoint(workspacePath, shadowRepoPath)

	// 尝试在未初始化的仓库中获取分支（应该失败）
	_, err = cp.Branch()
	if err == nil {
		t.Error("Branch should fail on uninitialized repo")
	}
}
