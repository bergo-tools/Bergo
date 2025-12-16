package berio

import (
	"bergo/utils/cli"
	"fmt"
	"sync"
)

type CliOutput struct {
	sync.Mutex
	llmPrinter *cli.LLMPrinter
	intents    map[string]func(string) string
}

func (l *CliOutput) initPrinter() {
	if l.llmPrinter == nil {
		l.llmPrinter = cli.NewLLMPrinter()
	}
}
func (l *CliOutput) OnLLMResponse(response string, isReasoning bool) {
	l.Lock()
	defer l.Unlock()
	if len(response) == 0 {
		return
	}
	l.initPrinter()
	if isReasoning {
		l.llmPrinter.Reasoning(response)
	} else {
		l.llmPrinter.Print(response)
	}
}

func (l *CliOutput) Stop() string {
	l.Lock()
	defer l.Unlock()
	if l.llmPrinter != nil {
		str := l.llmPrinter.Stop()
		l.llmPrinter = nil
		return str
	}
	return ""
}
func (l *CliOutput) UpdateTail(tail string) {
	l.Lock()
	defer l.Unlock()
	if l.llmPrinter == nil {
		l.initPrinter()
	}
	l.llmPrinter.UpdateTail(tail)
}

func (l *CliOutput) OnSystemMsg(msg interface{}, typ int) {
	l.Lock()
	defer l.Unlock()
	if l.llmPrinter != nil {
		l.llmPrinter.Stop()
		l.llmPrinter = nil
	}
	if typ == MsgTypeText {
		cli.PrintSystemText(msg.(string))
	}
	if typ == MsgTypeWarning {
		fmt.Println("")
		cli.PrintWarningText(msg.(string))
	}
	if typ == MsgTypeDump {
		fmt.Println(msg.(string))
	}
}

/*
	func (l *CliOutput) InjectHooks(tagFilter *utils.TagFiter) {
		render, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(pterm.GetTerminalWidth()*4/10-2))
		l.Lock()
		defer l.Unlock()
		intents := l.intents
		for tag, f := range intents {
			tagFilter.AddFinishHook(tag, func(content string) {
				l.Lock()
				defer l.Unlock()
				l.initPrinter()
				l.llmPrinter.DirectPrint(utils.ToolUseStyle(f(content)))
			})
			tagFilter.AddBeginHook(tag, func(content string) {
				l.Lock()
				defer l.Unlock()
				l.initPrinter()
				l.llmPrinter.UpdateTail(utils.ToolUseStyle(locales.Sprintf("Bergo is making tool calls...")))
			})
		}
		var fileTpl = `~~~%s

%s
~~~
`

		tagFilter.AddFinishHook("edit_diff", func(content string) {
			l.Lock()
			defer l.Unlock()
			l.initPrinter()
			tagfilter := utils.NewTagFiter("path", "search", "replace")
			tagfilter.Filter(content)
			tagfilter.Close()
			path := tagfilter.GetInnerConetent("path")
			searchContent := tagfilter.GetInnerConetent("search")
			replaceContent := tagfilter.GetInnerConetent("replace")

			ext := utils.GetLangByExt(path)
			searchContent = fmt.Sprintf(fileTpl, ext, searchContent)
			replaceContent = fmt.Sprintf(fileTpl, ext, replaceContent)
			searchContent, _ = render.Render(searchContent)
			replaceContent, _ = render.Render(replaceContent)
			l.llmPrinter.DirectPrint(utils.SearchReplaceStyle(path, searchContent, replaceContent))

		})
		tagFilter.AddFinishHook("edit_whole", func(content string) {
			l.Lock()
			defer l.Unlock()
			l.initPrinter()
			tagfilter := utils.NewTagFiter("path", "replace")
			tagfilter.Filter(content)
			tagfilter.Close()
			path := tagfilter.GetInnerConetent("path")
			replaceContent := tagfilter.GetInnerConetent("replace")

			ext := utils.GetLangByExt(path)
			replaceContent = fmt.Sprintf(fileTpl, ext, replaceContent)
			replaceContent, _ = render.Render(replaceContent)
			l.llmPrinter.DirectPrint(utils.SearchReplaceStyle(path, "", replaceContent))
		})
		tagFilter.AddFinishHook("stop_loop", func(content string) {
			l.Lock()
			defer l.Unlock()
			l.initPrinter()
			text, _ := render.Render(content)
			l.llmPrinter.DirectPrint(utils.StopLoopMessageStyle(text))
		})
	}
*/
type CliInput struct {
	options cli.InputOptions
}

func NewCliInput(options cli.InputOptions) BerInput {
	return &CliInput{
		options: options,
	}
}
func (r *CliInput) Read() (string, error) {
	if r.options.Multiline {
		return cli.NewTeaInput().WithHeader(r.options.String()).ReadMultilines()
	}
	return cli.NewTeaInput().WithHeader(r.options.String()).Read()
}
func (r *CliInput) Select(prompt string, options []string) string {
	return cli.CliSelect(prompt, options, 0)
}
func NewCliOutput() BerOutput {
	return &CliOutput{}
}
