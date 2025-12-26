package utils

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

type ReadFile struct {
	Path        string
	LineBudget  int
	WithLineNum bool
}

func (r *ReadFile) ReadFile() ([]string, error) {
	isBinary, err := IsBinaryFile(r.Path)
	if err != nil {
		return nil, err
	}
	if isBinary {
		return nil, fmt.Errorf("%s is a binary file", r.Path)
	}
	file, err := os.Open(r.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var result []string
	lineNum := 1
	for scanner.Scan() {
		if r.LineBudget > 0 && lineNum > r.LineBudget {
			result = append(result, fmt.Sprintf("...content after %d lines are truncated...\n", r.LineBudget))
			break
		}
		line := scanner.Text()
		if r.WithLineNum {
			result = append(result, fmt.Sprintf("%d|%s\n", lineNum, line))
		} else {
			result = append(result, line+"\n")
		}
		lineNum++
	}
	return result, nil
}

func (r *ReadFile) ReadFileTruncated(start int, end int) ([]string, error) {
	isBinary, err := IsBinaryFile(r.Path)
	if err != nil {
		return nil, err
	}
	if isBinary {
		return nil, fmt.Errorf("%s is a binary file", r.Path)
	}
	//offset and end are line numbers
	file, err := os.Open(r.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var result []string
	lineNum := 1
	if end == 0 {
		end = math.MaxInt
	}
	for scanner.Scan() {
		if lineNum >= start && lineNum <= end {
			line := scanner.Text()
			if r.WithLineNum {
				result = append(result, fmt.Sprintf("%d|%s\n", lineNum, line))
			} else {
				result = append(result, line+"\n")
			}
		}
		lineNum++
	}
	return result, nil
}

func (r *ReadFile) ReadFileWhole() (string, error) {
	isBinary, err := IsBinaryFile(r.Path)
	if err != nil {
		return "", err
	}
	if isBinary {
		return "", fmt.Errorf("%s is a binary file", r.Path)
	}
	data, err := os.ReadFile(r.Path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsBinaryFile 检测文件是否为二进制文件
// 通过读取文件前8KB内容，检查是否包含NULL字节或过多的非文本字符
func IsBinaryFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// 读取前8KB内容进行检测
	buf := make([]byte, 8192)
	n, err := file.Read(buf)
	if n == 0 {
		// 空文件视为文本文件
		return false, nil
	}
	buf = buf[:n]

	// 检测NULL字节和统计非文本字符
	nonTextCount := 0
	for _, b := range buf {
		// NULL字节是二进制文件的典型特征
		if b == 0 {
			return true, nil
		}
		// 统计非文本字符（控制字符，排除换行、回车、制表符）
		if b < 32 && b != '\n' && b != '\r' && b != '\t' {
			nonTextCount++
		}
	}

	// 如果非文本字符超过30%，认为是二进制文件
	if float64(nonTextCount)/float64(len(buf)) > 0.3 {
		return true, nil
	}

	return false, nil
}
