package test

import (
	"strings"
	"testing"

	"bergo/utils"
)

func TestCheckSyntaxError_UnsupportedFileType(t *testing.T) {
	filename := "test.unsupported"
	content := []byte("any content")
	err := utils.CheckSyntaxError(filename, content)
	if err == nil {
		t.Fatal("expected error for unsupported file type, got nil")
	}
	expectedErr := "unsupported file type: .unsupported"
	if err.Error() != expectedErr {
		t.Errorf("expected error: %q, got: %q", expectedErr, err.Error())
	}
}

func TestCheckSyntaxError_ValidGoCode(t *testing.T) {
	filename := "valid.go"
	content := []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}")
	err := utils.CheckSyntaxError(filename, content)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestCheckSyntaxError_InvalidGoCode(t *testing.T) {
	filename := "invalid.go"
	content := []byte("package main\n\nfunc main() {") // 缺少闭合括号
	err := utils.CheckSyntaxError(filename, content)
	t.Log(err)
	if err == nil {
		t.Fatal("expected syntax error, got nil")
	}

	// 验证错误信息格式
	expectedStart := "syntax errors found:"
	if !strings.Contains(err.Error(), expectedStart) {
		t.Errorf("error should start with %q, got: %q", expectedStart, err.Error())
	}

	// 验证包含具体错误位置
	expectedLocation := "Line 3:"
	if !strings.Contains(err.Error(), expectedLocation) {
		t.Errorf("error should contain location %q, got: %q", expectedLocation, err.Error())
	}
}

func TestCheckSyntaxError_ValidJavaScript(t *testing.T) {
	filename := "valid.js"
	content := []byte("function test() { return 1; }")
	err := utils.CheckSyntaxError(filename, content)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestCheckSyntaxError_InvalidJavaScript(t *testing.T) {
	filename := "invalid.js"
	content := []byte("function test() {baba = ;}") // 缺少闭合括号
	err := utils.CheckSyntaxError(filename, content)
	t.Log(err)
	if err == nil {
		t.Fatal("expected syntax error, got nil")
	}
}
