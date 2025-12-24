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
