package cli

import (
	"bergo/config"
	"bergo/locales"
	"bergo/utils"
	"sort"
	"strings"
	"sync"
)

var nowPath *utils.NowPath = utils.NewNowPath()

var cmdsuggestions = []CompletionItem{
	{Text: "/exit", Description: locales.Sprintf("exit bergo")},
	{Text: "/help", Description: locales.Sprintf("instructions about bergo")},
	{Text: "/view", Description: locales.Sprintf("switch to view mode")},
	{Text: "/planner", Description: locales.Sprintf("switch to planner mode")},
	{Text: "/agent", Description: locales.Sprintf("switch to agent mode")},
	{Text: "/multiline", Description: locales.Sprintf("switch to multiline input mode, ctrl+C to exit")},
	{Text: "/revert", Description: locales.Sprintf("revert to last checkpoint")},
	{Text: "/history", Description: locales.Sprintf("show timeline viewer")},
	{Text: "/sessions", Description: locales.Sprintf("show sessions viewer")},
	{Text: "/clear", Description: locales.Sprintf("clear everthing. start a new session")},
	{Text: "/model", Description: locales.Sprintf("switch model")},
	{Text: "/compact", Description: locales.Sprintf("compact the context")},
}

func getCompletion(prefix string, whole string) string {
	if prefix == "" {
		return whole
	}
	if strings.HasPrefix(whole, prefix) {
		return whole[len(prefix):]
	}
	return ""
}
func wordAfterWhite(input string) string {
	strs := strings.Split(input, " ")
	for i := len(strs) - 1; i >= 0; i-- {
		if strings.TrimSpace(strs[i]) != "" {
			return strs[i]
		}
	}
	return ""
}

func getCompletionItems(prefix string) []CompletionItem {
	files := nowPath.MatchFiles(prefix)
	options := []CompletionItem{}
	for _, file := range files {
		options = append(options, CompletionItem{Text: file.Path, Description: file.Name})
	}
	return options

}

var once sync.Once
var modelTrie *utils.Trie
var atTrie *utils.Trie
var AtCmds = []struct {
	Text string
	Gen  func() string
}{
	{Text: "@file:", Gen: func() string { return locales.Sprintf("file path of a file or directory") }},
	//{Text: "@web ", Gen: func() string { return locales.Sprintf("use web search") }},
}

// GetAtCmds 返回当前可用的@命令列表，根据模型配置动态决定
func GetAtCmds(modelIdentifier string) []struct {
	Text string
	Gen  func() string
} {
	cmds := make([]struct {
		Text string
		Gen  func() string
	}, len(AtCmds))
	copy(cmds, AtCmds)

	// 检查当前模型是否支持vision
	modelConfig := config.GlobalConfig.GetModelConfig(modelIdentifier)
	if modelConfig != nil && modelConfig.SupportVision {
		cmds = append(cmds, struct {
			Text string
			Gen  func() string
		}{Text: "@img:", Gen: func() string { return locales.Sprintf("file path of an image") }})
	}
	return cmds
}

func BergoCompleter(input string, cusorPos int) []*CompletionItem {
	result := BergoCompleterRaw(input, cusorPos)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Text < result[j].Text
	})
	return result
}
func BergoCompleterRaw(inputStr string, cusorPos int) []*CompletionItem {
	once.Do(func() {
		modelTrie = utils.NewTrie()
		for _, model := range config.GlobalConfig.Models {
			modelTrie.Put(model.Identifier, 1)
		}
	})
	// 动态构建atTrie，根据当前模型配置
	atTrie = utils.NewTrie()
	for _, cmd := range GetAtCmds(config.GlobalConfig.MainModel) {
		atTrie.Put(cmd.Text, cmd.Gen)
	}
	input := []rune(inputStr)
	// 过滤匹配的补全项
	var result []*CompletionItem
	lastAtPos := -1
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == '@' {
			lastAtPos = i
			break
		}
	}
	if cusorPos <= lastAtPos {
		return result
	}
	if lastAtPos >= 0 {
		atCmd := string(input[lastAtPos:cusorPos])
		if strings.HasPrefix(atCmd, "@file:") {
			after := strings.TrimPrefix(atCmd, "@file:")
			path := after
			if after == "" {
				path = "."
			}
			if strings.Contains(path, " ") {
				return result
			}
			nowPath.Update(path)
			pathItem := getCompletionItems(path)
			for _, item := range pathItem {
				completion := getCompletion(after, item.Text)
				result = append(result, &CompletionItem{Text: item.Text, Description: item.Description, Completion: completion, AddSapce: true})
			}
		} else if strings.HasPrefix(atCmd, "@img:") {
			modelConfig := config.GlobalConfig.GetModelConfig(config.GlobalConfig.MainModel)
			if modelConfig == nil || !modelConfig.SupportVision {
				return result
			}
			after := strings.TrimPrefix(atCmd, "@img:")
			path := after
			if after == "" {
				path = "."
			}
			if strings.Contains(path, " ") {
				return result
			}
			nowPath.Update(path)
			files := nowPath.MatchFiles(path)
			for _, file := range files {
				// 只补全图片文件和目录
				isDir := strings.HasSuffix(file.Name, "/")
				if isDir || utils.IsImageFile(file.Path) {
					completion := getCompletion(after, file.Path)
					result = append(result, &CompletionItem{Text: file.Path, Description: file.Name, Completion: completion, AddSapce: true})
				}
			}
		} else {
			atTrie.WalkPath(atCmd, func(key string, value interface{}) {
				desc := value.(func() string)()
				completion := getCompletion(atCmd, key)
				result = append(result, &CompletionItem{Text: key, Description: desc, Completion: completion})
			})
		}
		if len(result) == 1 && result[0].Completion == "" {
			return nil
		}
		return result
	}

	// 检查是否是命令补全
	if strings.HasPrefix(inputStr, "/") {
		for _, s := range cmdsuggestions {
			if strings.HasPrefix(s.Text, inputStr) {
				result = append(result, &CompletionItem{Text: s.Text, Description: s.Description, Completion: getCompletion(inputStr, s.Text)})
			}
		}
	}

	if strings.HasPrefix(inputStr, "/model ") {
		prefix := wordAfterWhite(inputStr)
		if prefix == "/model" || prefix == "" {
			prefix = ""
		}
		modelTrie.WalkPath(prefix, func(key string, value interface{}) {
			result = append(result, &CompletionItem{Text: key, Description: config.GlobalConfig.GetModelConfig(key).Provider, Completion: getCompletion(prefix, key)})
		})

	}

	//完全匹配就不显示菜单了
	if len(result) == 1 && result[0].Completion == "" {
		return nil
	}
	return result
}
