package utils

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/cue"
	"github.com/smacker/go-tree-sitter/dockerfile"
	"github.com/smacker/go-tree-sitter/elixir"
	"github.com/smacker/go-tree-sitter/elm"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/groovy"
	"github.com/smacker/go-tree-sitter/hcl"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/ocaml"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/protobuf"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/sql"
	"github.com/smacker/go-tree-sitter/svelte"
	"github.com/smacker/go-tree-sitter/swift"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

// 扩展名到tree-sitter语言的映射
var langMap = map[string]*sitter.Language{
	".bash":       bash.GetLanguage(),
	".c":          c.GetLanguage(),
	".cpp":        cpp.GetLanguage(),
	".cs":         csharp.GetLanguage(),
	".css":        css.GetLanguage(),
	".cue":        cue.GetLanguage(),
	".dockerfile": dockerfile.GetLanguage(),
	".ex":         elixir.GetLanguage(),
	".exs":        elixir.GetLanguage(),
	".elm":        elm.GetLanguage(),
	".go":         golang.GetLanguage(),
	".groovy":     groovy.GetLanguage(),
	".h":          c.GetLanguage(),
	".hpp":        cpp.GetLanguage(),
	".hcl":        hcl.GetLanguage(),
	".html":       html.GetLanguage(),
	".java":       java.GetLanguage(),
	".js":         javascript.GetLanguage(),
	".kt":         kotlin.GetLanguage(),
	".kts":        kotlin.GetLanguage(),
	".lua":        lua.GetLanguage(),
	".ml":         ocaml.GetLanguage(),
	".mli":        ocaml.GetLanguage(),
	".php":        php.GetLanguage(),
	".proto":      protobuf.GetLanguage(),
	".py":         python.GetLanguage(),
	".rb":         ruby.GetLanguage(),
	".rs":         rust.GetLanguage(),
	".scala":      scala.GetLanguage(),
	".sc":         scala.GetLanguage(),
	".sh":         bash.GetLanguage(),
	".sql":        sql.GetLanguage(),
	".svelte":     svelte.GetLanguage(),
	".swift":      swift.GetLanguage(),
	".toml":       toml.GetLanguage(),
	".ts":         typescript.GetLanguage(),
	".tsx":        tsx.GetLanguage(),
	".yaml":       yaml.GetLanguage(),
	".yml":        yaml.GetLanguage(),
	"Dockerfile":  dockerfile.GetLanguage(),
}

var extLangStringMap = map[string]string{
	".bash":       "bash",
	".c":          "c",
	".cpp":        "cpp",
	".cs":         "csharp",
	".css":        "css",
	".cue":        "cue",
	".dockerfile": "dockerfile",
	".ex":         "elixir",
	".exs":        "elixir",
	".elm":        "elm",
	".go":         "go",
	".groovy":     "groovy",
	".h":          "c",
	".hpp":        "cpp",
	".hcl":        "hcl",
	".html":       "html",
	".java":       "java",
	".js":         "javascript",
	".kt":         "kotlin",
	".kts":        "kotlin",
	".lua":        "lua",
	".ml":         "ocaml",
	".mli":        "ocaml",
	".php":        "php",
	".proto":      "protobuffer",
	".py":         "python",
	".rb":         "ruby",
	".rs":         "rust",
	".scala":      "scala",
	".sc":         "scala",
	".sh":         "bash",
	".sql":        "sql",
	".svelte":     "svelte",
	".swift":      "swift",
	".toml":       "toml",
	".ts":         "typescript",
	".tsx":        "tsx",
	".yaml":       "yaml",
	".yml":        "yaml",
	"Dockerfile":  "dockerfile",
}

func GetLangByExt(path string) string {
	return extLangStringMap[strings.ToLower(filepath.Ext(path))]
}

func IsFileSupported(filename string) bool {
	_, ok := langMap[strings.ToLower(filepath.Ext(filename))]
	return ok
}

func CheckSyntaxError(filename string, content []byte) error {
	// 获取文件扩展名并转换为小写
	ext := strings.ToLower(filepath.Ext(filename))

	// 检查是否支持该文件类型
	lang, ok := langMap[ext]
	if !ok {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	// 创建解析器并设置语言
	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	// 解析代码
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}
	defer tree.Close()

	// 检查语法错误
	rootNode := tree.RootNode()
	if rootNode.HasError() {
		// 收集所有错误节点
		query, err := sitter.NewQuery([]byte("(ERROR) @err"), lang)
		if err != nil {
			return fmt.Errorf("query creation failed: %w", err)
		}
		defer query.Close()

		cursor := sitter.NewQueryCursor()
		defer cursor.Close()
		cursor.Exec(query, rootNode)

		var errors []string
		for {
			match, ok := cursor.NextMatch()
			if !ok {
				break
			}
			for _, capture := range match.Captures {
				node := capture.Node
				start := node.StartPoint()
				end := node.EndPoint()
				errorContent := string(content[node.StartByte():node.EndByte()])

				errors = append(errors, fmt.Sprintf(
					"Line %d:%d - %d:%d: %s",
					start.Row+1, start.Column+1,
					end.Row+1, end.Column+1,
					errorContent,
				))
			}
		}

		if len(errors) > 0 {
			return fmt.Errorf("syntax errors found:\n%s", strings.Join(errors, "\n"))
		}
	}

	return nil
}
