package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var workspace = "../../"
var outputDir = "../lang"

//go:generate go run extract.go
func main() {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// 创建一个map来存储提取的字符串
	localeStrings := make(map[string]bool)

	// 遍历工作区中的所有Go文件
	err := filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理.go文件，但排除测试文件和当前extract.go文件
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") &&
			!strings.HasSuffix(info.Name(), "_test.go") &&
			!strings.Contains(path, "locales/extract") {
			// 解析Go源文件
			strings, err := extractLocaleStrings(path)
			if err != nil {
				fmt.Printf("Error parsing %s: %v\n", path, err)
				return nil
			}

			// 将提取的字符串添加到map中
			for _, s := range strings {
				localeStrings[s] = true
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking workspace: %v\n", err)
		return
	}

	// 生成JSON文件
	if err := generateJSONFiles(localeStrings); err != nil {
		fmt.Printf("Error generating JSON files: %v\n", err)
		return
	}

	fmt.Printf("Successfully generated JSON files in %s\n", outputDir)
}

// extractLocaleStrings 解析Go源文件并提取locales包函数调用中的字符串
func extractLocaleStrings(filename string) ([]string, error) {
	// 创建token文件集
	fset := token.NewFileSet()

	// 解析源文件
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err
	}

	var result []string

	// 遍历AST节点
	ast.Inspect(f, func(n ast.Node) bool {
		// 查找函数调用表达式
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// 检查是否是locales包的函数调用
		funcName := getFunctionName(callExpr)
		if funcName == "locales.Sprintf" || funcName == "locales.Sprint" || funcName == "locales.Errorf" {
			// 提取第一个参数（字符串参数）
			if len(callExpr.Args) > 0 {
				if basicLit, ok := callExpr.Args[0].(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
					// 使用strconv.Unquote正确解析字符串字面量
					strValue, err := strconv.Unquote(basicLit.Value)
					if err != nil {
						// 如果解析失败，回退到简单处理
						strValue = strings.Trim(basicLit.Value, "\"")
					}
					result = append(result, strValue)
				}
			}
		}

		return true
	})

	return result, nil
}

// getFunctionName 获取函数调用的完整名称
func getFunctionName(callExpr *ast.CallExpr) string {
	switch fun := callExpr.Fun.(type) {
	case *ast.SelectorExpr:
		// 处理包.函数形式的调用
		if ident, ok := fun.X.(*ast.Ident); ok {
			return ident.Name + "." + fun.Sel.Name
		}
	case *ast.Ident:
		// 处理本地函数调用
		return fun.Name
	}
	return ""
}

// generateJSONFiles 生成JSON翻译文件
func generateJSONFiles(localeStrings map[string]bool) error {
	// 定义要生成的语言文件
	languages := []string{"zh"} // 可以根据需要添加更多语言

	for _, lang := range languages {
		langFile := filepath.Join(outputDir, lang+".json")

		// 读取现有的翻译文件（如果存在）
		existingTranslations := make(map[string]string)
		if data, err := os.ReadFile(langFile); err == nil {
			if err := json.Unmarshal(data, &existingTranslations); err != nil {
				fmt.Printf("Warning: Failed to parse existing %s.json: %v\n", lang, err)
			}
		}

		// 创建新的翻译map
		newTranslations := make(map[string]string)

		// 为每个提取的字符串添加翻译
		for str := range localeStrings {
			if translation, exists := existingTranslations[str]; exists && translation != "" {
				// 保留现有的非空翻译
				newTranslations[str] = translation
			} else {
				// 新字符串或空翻译，使用空字符串
				newTranslations[str] = ""
			}
		}

		// 将翻译map转换为JSON
		jsonData, err := json.MarshalIndent(newTranslations, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal %s.json: %v", lang, err)
		}

		// 写入JSON文件
		if err := os.WriteFile(langFile, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write %s.json: %v", lang, err)
		}

		fmt.Printf("Generated %s\n", langFile)
	}

	return nil
}
