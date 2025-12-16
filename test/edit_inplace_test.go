package test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bergo/utils"
)

// 创建临时测试文件
func createTestFile(t *testing.T, content string) string {
	t.Helper()
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "edit_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	// 确保测试结束后清理临时目录
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	// 创建临时文件
	filePath := filepath.Join(tempDir, "test_file.txt")
	// 确保内容不包含末尾换行符
	content = strings.TrimSuffix(content, "\n")
	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	return filePath
}

// 读取文件内容
func readFileContent(t *testing.T, filePath string) string {
	t.Helper()
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	return string(content)
}

// TestEditInplaceNormal 测试正常编辑场景
func TestEditInplaceNormal(t *testing.T) {
	t.Parallel()

	// 准备测试数据
	originalContent := "a\nb\nc\nd\ne"
	filePath := createTestFile(t, originalContent)

	// 初始化 Edit 对象
	edit := &utils.Edit{
		Path: filePath,
	}

	// 读取文件原始内容来获取实际的行内容（包括换行符）
	rf := &utils.ReadFile{
		Path:        filePath,
		LineBudget:  10,
		WithLineNum: false,
	}
	lines, err := rf.ReadFile()
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	// 确保我们有足够的行进行测试
	if len(lines) < 5 {
		t.Fatalf("文件行数不足，期望至少5行，实际有%d行", len(lines))
	}

	// 测试用例1: 替换多行，使用实际读取的行内容
	replaceContent := "B\n"
	err = edit.EditInplace(2, 3, lines[1], lines[2], replaceContent)
	if err != nil {
		t.Fatalf("替换行失败: %v", err)
	}

	expectedContent := "a\nB\nd\ne\n"
	actualContent := readFileContent(t, filePath)
	if actualContent != expectedContent {
		t.Errorf("替换行内容不匹配\n期望: %s\n实际: %s", expectedContent, actualContent)
	}
}

// TestEditInplaceEdgeCases 测试边界条件和错误情况
func TestEditInplaceEdgeCases(t *testing.T) {
	t.Parallel()

	// 准备测试数据
	originalContent := "a\nb\nc"
	filePath := createTestFile(t, originalContent)

	// 初始化 Edit 对象
	edit := &utils.Edit{
		Path: filePath,
	}

	// 读取文件原始内容来获取实际的行内容
	rf := &utils.ReadFile{
		Path:        filePath,
		LineBudget:  10,
		WithLineNum: false,
	}
	lines, err := rf.ReadFile()
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	// 确保我们有足够的行进行测试
	if len(lines) < 3 {
		t.Fatalf("文件行数不足，期望至少3行，实际有%d行", len(lines))
	}

	// 获取实际行内容（去除换行符）
	aLine := strings.TrimSuffix(lines[0], "\n")
	bLine := strings.TrimSuffix(lines[1], "\n")
	cLine := strings.TrimSuffix(lines[2], "\n")

	// 测试用例1: 无效参数 (start >= end)
	err = edit.EditInplace(3, 2, cLine, bLine, "replacement")
	if err == nil {
		t.Error("期望参数无效错误，但未返回错误")
	} else if !strings.Contains(err.Error(), "start line") || !strings.Contains(err.Error(), "must be less than end line") {
		t.Errorf("期望错误消息包含 'start line' 和 'must be less than end line'，但得到: %s", err.Error())
	}

	// 测试用例2: 无效参数 (start == end)
	err = edit.EditInplace(2, 2, bLine, bLine, "replacement")
	if err == nil {
		t.Error("期望参数无效错误，但未返回错误")
	} else if !strings.Contains(err.Error(), "start line") || !strings.Contains(err.Error(), "must be less than end line") {
		t.Errorf("期望错误消息包含 'start line' 和 'must be less than end line'，但得到: %s", err.Error())
	}

	// 重置文件内容
	ioutil.WriteFile(filePath, []byte(originalContent), 0644)

	// 测试用例3: 文件行数不足
	err = edit.EditInplace(1, 5, aLine, "e", "replacement")
	if err == nil {
		t.Error("期望文件行数不足错误，但未返回错误")
	} else if !strings.Contains(err.Error(), "end line") || !strings.Contains(err.Error(), "must be less than file line count") {
		t.Errorf("期望错误消息包含 'end line' 和 'must be less than file line count'，但得到: %s", err.Error())
	}

	// 测试用例4: 起始行内容不匹配
	err = edit.EditInplace(2, 3, "wrong", cLine, "replacement")
	if err == nil {
		t.Error("期望起始行内容不匹配错误，但未返回错误")
	} else if !strings.Contains(err.Error(), "start_line") {
		t.Errorf("期望错误消息包含 'start_line'，但得到: %s", err.Error())
	}

	// 测试用例5: 结束行内容不匹配
	err = edit.EditInplace(1, 2, aLine, "wrong", "replacement")
	if err == nil {
		t.Error("期望结束行内容不匹配错误，但未返回错误")
	} else if !strings.Contains(err.Error(), "end_line") {
		t.Errorf("期望错误消息包含 'end_line'，但得到: %s", err.Error())
	}
}

// TestEditInplaceSpecialCase 测试特殊情况 (start != 0 && end == 0)
func TestEditInplaceSpecialCase(t *testing.T) {
	t.Parallel()

	// 准备测试数据
	originalContent := "line 1\nline 2\nline 3"
	filePath := createTestFile(t, originalContent)

	// 初始化 Edit 对象
	edit := &utils.Edit{
		Path: filePath,
	}

	// 读取文件原始内容来获取实际的行内容
	rf := &utils.ReadFile{
		Path:        filePath,
		LineBudget:  10,
		WithLineNum: false,
	}
	lines, err := rf.ReadFile()
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	// 确保我们有足够的行进行测试
	if len(lines) < 3 {
		t.Fatalf("文件行数不足，期望至少3行，实际有%d行", len(lines))
	}

	// 获取实际行内容（去除换行符）
	line2 := strings.TrimSuffix(lines[1], "\n")

	// 测试特殊情况: start != 0 && end == 0
	// 根据代码实现，这种情况下应该设置 end = start, end_line = start_line
	// 但由于 start == end 不被允许，所以期望返回错误
	replaceContent := "modified line 2"
	err = edit.EditInplace(2, 0, line2, line2, replaceContent)
	if err == nil {
		t.Error("期望参数无效错误，但未返回错误")
	} else if !strings.Contains(err.Error(), "start line") || !strings.Contains(err.Error(), "must be less than end line") {
		t.Errorf("期望错误消息包含 'start line' 和 'must be less than end line'，但得到: %s", err.Error())
	}
}
