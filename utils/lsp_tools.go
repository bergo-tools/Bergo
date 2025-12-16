package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bergo/lsp"
)

type LspTool struct {
	Lang string
}

type LspLine struct {
	Line    int
	Content string
}

type LspResultItem struct {
	FilePath string //如果在当前目录下，取想对路径
	Lines    []LspLine
}

/*
filePath 文件路径,可能是相对也可能是绝对路径，需要取绝对路径
line 行号
symbol 需要查询的符号，如果一行有多个该符号就取第一个

流程：
1. 取绝对路径
2. 调用LSP服务器查询引用
3. 解析LSP服务器返回结果，同一文件当中的引用合并到同一个LspResultItem中
4. 返回引用结果
*/
func (t *LspTool) FindReference(filePath string, line int, symbol string) ([]*LspResultItem, error) {
	// 获取LSP客户端函数
	clientFunc, ok := lsp.GetClientFunc(t.Lang)
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", t.Lang)
	}

	// 创建LSP客户端
	client, err := clientFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to create LSP client: %w", err)
	}
	defer client.Shutdown(context.Background())

	// 获取绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 计算符号在当前行的字符位置（简单实现：查找符号在行中的位置）
	character, err := t.findSymbolCharacter(absPath, line, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to find symbol character position: %w", err)
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 查询引用
	locations, err := client.FindReference(ctx, absPath, line-1, character) // LSP行号从0开始
	if err != nil {
		return nil, fmt.Errorf("failed to find references: %w", err)
	}

	// 解析和合并结果
	return t.parseLocations(locations)
}

// findSymbolCharacter 查找符号在行中的字符位置
func (t *LspTool) findSymbolCharacter(filePath string, line int, symbol string) (int, error) {
	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	if line-1 < 0 || line-1 >= len(lines) {
		return 0, fmt.Errorf("line number %d out of range", line)
	}

	// 查找符号在当前行的位置
	lineContent := lines[line-1]
	index := strings.Index(lineContent, symbol)
	if index == -1 {
		return 0, fmt.Errorf("symbol '%s' not found in line %d", symbol, line)
	}

	return index, nil
}

// parseLocations 解析LSP位置结果，合并同一文件的结果
func (t *LspTool) parseLocations(locations []lsp.Location) ([]*LspResultItem, error) {
	if len(locations) == 0 {
		return []*LspResultItem{}, nil
	}

	// 按文件分组位置
	fileGroups := make(map[string][]lsp.Location)
	for _, loc := range locations {
		filePath := strings.TrimPrefix(loc.URI, "file://")
		fileGroups[filePath] = append(fileGroups[filePath], loc)
	}

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = ""
	}

	var result []*LspResultItem
	for filePath, locs := range fileGroups {
		// 读取文件内容
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue // 如果无法读取文件，跳过
		}

		lines := strings.Split(string(content), "\n")

		// 创建结果项
		item := &LspResultItem{
			FilePath: filePath,
			Lines:    []LspLine{},
		}

		// 使用相对路径如果可能
		if currentDir != "" {
			if relPath, err := filepath.Rel(currentDir, filePath); err == nil {
				item.FilePath = relPath
			}
		}

		// 收集所有相关的行
		lineMap := make(map[int]bool)
		for _, loc := range locs {
			startLine := loc.Range.Start.Line
			endLine := loc.Range.End.Line

			for lineNum := startLine; lineNum <= endLine; lineNum++ {
				lineMap[lineNum] = true
			}
		}

		// 按行号排序并添加到结果中
		for lineNum := range lineMap {
			if lineNum >= 0 && lineNum < len(lines) {
				item.Lines = append(item.Lines, LspLine{
					Line:    lineNum + 1, // 转换为1基数的行号
					Content: lines[lineNum],
				})
			}
		}

		// 按行号排序
		for i := 0; i < len(item.Lines); i++ {
			for j := i + 1; j < len(item.Lines); j++ {
				if item.Lines[i].Line > item.Lines[j].Line {
					item.Lines[i], item.Lines[j] = item.Lines[j], item.Lines[i]
				}
			}
		}

		if len(item.Lines) > 0 {
			result = append(result, item)
		}
	}

	return result, nil
}

/*
filePath 文件路径,可能是相对也可能是绝对路径，需要取绝对路径
line 行号
symbol 需要查询的符号，如果一行有多个该符号就取第一个

流程：
1. 取绝对路径
2. 调用LSP服务器查询定义
3. 解析LSP服务器返回结果，同一文件当中的定义合并到同一个LspResultItem中
4. 返回定义结果
*/
func (t *LspTool) FindDefinition(filePath string, line int, symbol string) ([]*LspResultItem, error) {
	// 获取LSP客户端函数
	clientFunc, ok := lsp.GetClientFunc(t.Lang)
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", t.Lang)
	}

	// 创建LSP客户端
	client, err := clientFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to create LSP client: %w", err)
	}
	defer client.Shutdown(context.Background())

	// 获取绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 计算符号在当前行的字符位置（简单实现：查找符号在行中的位置）
	character, err := t.findSymbolCharacter(absPath, line, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to find symbol character position: %w", err)
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 查询定义
	locations, err := client.FindDefinition(ctx, absPath, line-1, character) // LSP行号从0开始
	if err != nil {
		return nil, fmt.Errorf("failed to find definitions: %w", err)
	}

	// 解析和合并结果
	return t.parseLocations(locations)
}
