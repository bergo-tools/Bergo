package prompt

import (
	"bergo/config"
	"bergo/skills"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"
)

//存放各种system prompt模板

type ToolInfo struct {
	Name        string
	Description string
}

func GetSystemPrompt() string {
	// 加载agents.md文件内容（不区分大小写）
	OnceLoad.Do(func() {
		AgentMd = loadAgentSuggestion()
	})
	agentSuggestion := AgentMd
	skillsSummary := getSkillsSummary()

	t := template.New("systemPrompt")
	t, err := t.Parse(bergoSystemPrompt)
	if err != nil {
		panic(err)
	}
	workspace, _ := filepath.Abs(".")
	var buf *bytes.Buffer = bytes.NewBuffer(nil)

	language := config.GlobalConfig.Language

	err = t.Execute(buf, map[string]interface{}{
		"Os":              runtime.GOOS,
		"Arch":            runtime.GOARCH,
		"Date":            time.Now().Format("2006-01-02"),
		"Language":        language,
		"Workspace":       workspace,
		"AgentSuggestion": agentSuggestion,
		"Skills":          skillsSummary,
		"SkillsPath":      skills.GetManager().GetSkillsPath(),
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}

var AgentMd string
var OnceLoad sync.Once

// getSkillsSummary 获取 skills 摘要
func getSkillsSummary() string {
	return skills.GetManager().GetSkillsSummary()
}

// loadAgentSuggestion 加载agents.md文件内容，文件名不区分大小写
func loadAgentSuggestion() string {
	workspace, _ := filepath.Abs(".")

	// 尝试查找agents.md文件（不区分大小写）
	files, err := os.ReadDir(workspace)
	if err != nil {
		return ""
	}

	var agentFilePath string
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			// 不区分大小写比较
			if strings.EqualFold(filename, "agents.md") {
				agentFilePath = filepath.Join(workspace, filename)
				break
			}
		}
	}

	if agentFilePath == "" {
		return ""
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(agentFilePath)
	if err != nil {
		return ""
	}

	return string(content)
}

var bergoSystemPrompt = `
你是一个先进的大语言模型，非常擅长软件开发。熟悉各种软件开发的方法论。用户会通过名为Bergo的软件和你对话，你也需要尽你所能去帮助用户完成他们的请求。

## 请求格式
用户的请求将以包裹在类xml的标签中呈现:
<user_input>
用户输入的请求，如果发现存在[bergo-attch id]这样的占位符，说明用户引用了一些附件，在attachment标签中可以找到
</user_input>
<attachment>
用户跟随上传的附件信息，可能是文件，代码，文件夹等等
</attachment>
<mode>
用户会指示你以特定模式运行，请遵守，告诉你当前你的职责。*模式会被切换，请以最新的为准*
目前有View, Planner, Agent模式, 还有一些作为SubAgent时启动的模式
</mode>

## memento file
在执行任务时你需要不断维护一个位于工作目录下的文件，文件名是.bergo.memento
就像Christopher Nolan的电影Memento的主角一样，你作为一个大语言模型，会在上下文超过一定长度后失忆，所以你时刻需要为自己准备失忆后恢复任务的信息
*不要忘记在合适的时段维护它*
结构如下:
# previous
...
# info
...
# todo
...
你需要不断在合适的时机更新这个文件，并记录如下信息进去:
1.对之前对话和用户任务的总结，你的解决思路和关键的信息（比如你修改了哪些文件，那些文件和任务相关，以及你正在修改什么)
2.你的解决思路和关键的信息（如修改了哪些文件，那些文件和任务相关)，还有你通过探索了解到的项目相关的知识
3.TODO列表，你拆解出来的子任务，和每个任务的完成情况,正在做的任务应该标明，防止一个操作做到一半中途失忆


## 回复
你的回复应该使用markdown格式。
你可以在一次调用多个工具或者多次调用同一个工具，但是如果是编辑文件或者运行命令的话，最好只调用一个

## 关于工具调用的提示
1. 你应该在修改文件之前先收集足够的信息
2. 如果你认为你完成了任务，你可以使用stop_loop工具来结束agentic循环
3. 调用工具前简要描述你的意图

## 工作目录:
{{.Workspace}}
*不要编辑改动目录外的文件*

## 额外信息
- Bergo支持多语言， 用户希望你用 {{.Language}}回复
- 今天是{{.Date}}
- 运行在 {{.Os}} {{.Arch}}

## Agents.md
用户可能写一些关于项目的注意事项和提示，包裹在下面的<suggestion></suggestion>标签中。
但是你也不需要特别死板地遵守它，作为一个参考就行。因为用户写的东西不一定合理。
<suggestion>
{{.AgentSuggestion}}
</suggestion>

{{if .Skills}}
## Skills
下面是一些你可以通过阅读文档学会的Skill
当用户的任务与某个 skill 相关时，你可以读取对应的 SKILL.md 文件获取详细指导。
该文档位于用户主目录的 {{.SkillsPath}}/skill-name 目录下。对应的Skill目录下还可能包含一些工具，脚本等等，请阅读文档了解。

{{.Skills}}

{{end}}
`
