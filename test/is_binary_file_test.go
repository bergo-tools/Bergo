package test

import (
	"os"
	"path/filepath"
	"testing"

	"bergo/utils"
)

func TestIsBinaryFile(t *testing.T) {
	testCases := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "plain text file",
			content:  []byte("Hello, World!\nThis is a text file.\n"),
			expected: false,
		},
		{
			name:     "text with tabs and carriage returns",
			content:  []byte("Line1\tTabbed\r\nLine2\r\n"),
			expected: false,
		},
		{
			name:     "empty file",
			content:  []byte{},
			expected: false,
		},
		{
			name:     "empty file2",
			content:  []byte(" \n\n\n   "),
			expected: false,
		},
		{
			name:     "file with NULL byte",
			content:  []byte("Hello\x00World"),
			expected: true,
		},
		{
			name:     "binary content with many NULL bytes",
			content:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}, // PNG header
			expected: true,
		},
		{
			name:     "Go source code",
			content:  []byte("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n"),
			expected: false,
		},
		{
			name:     "JSON content",
			content:  []byte(`{"name": "test", "value": 123}`),
			expected: false,
		},
		{
			name:     "file with high ratio of control characters",
			content:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0B, 0x0C},
			expected: true,
		},
		{
			name:     "UTF-8 Chinese text",
			content:  []byte("你好，世界！这是一个中文文本文件。\n"),
			expected: false,
		},
		{
			name:     "mixed text with few control chars (below threshold)",
			content:  append([]byte("Normal text content here"), 0x01),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "testfile")
			err := os.WriteFile(tmpFile, tc.content, 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			result, err := utils.IsBinaryFile(tmpFile)
			if err != nil {
				t.Fatalf("IsBinaryFile returned error: %v", err)
			}

			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestIsBinaryFile_NonExistentFile(t *testing.T) {
	_, err := utils.IsBinaryFile("/nonexistent/path/to/file")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestIsBinaryFile_RealBinaryFiles(t *testing.T) {
	// 测试真实的二进制文件（如果存在）
	binaryPaths := []string{
		"/bin/ls",
		"/bin/cat",
	}

	for _, path := range binaryPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		t.Run(path, func(t *testing.T) {
			result, err := utils.IsBinaryFile(path)
			if err != nil {
				t.Fatalf("IsBinaryFile returned error: %v", err)
			}
			if !result {
				t.Errorf("expected %s to be detected as binary", path)
			}
		})
	}
}
