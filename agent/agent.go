package agent

import (
	"bergo/berio"
	"bergo/config"
	"bergo/llm"
	"bergo/locales"
	"bergo/prompt"
	"bergo/tools"
	"bergo/utils"
	"bergo/utils/cli"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Agent struct {
	waitForUser bool
	attachments []*utils.Attachment
	agentMode   string
	multiline   bool
	timeline    *utils.Timeline
	output      berio.BerOutput
	toolHandler map[string]func(ctx context.Context, input *tools.AgentInput) *tools.AgentOutput
	toolSchema  []*llm.ToolSchema
	cmdHandler  map[string]func(input string) (string, bool)
	ignore      *utils.Ignore
	stop        bool
	stats       utils.Stat
	allowMap    map[string]bool

	sessionId string

	InteruptNum int
}

func NewMainAgent() *Agent {
	return &Agent{
		toolHandler: make(map[string]func(ctx context.Context, input *tools.AgentInput) *tools.AgentOutput),
		cmdHandler:  make(map[string]func(input string) (string, bool)),
	}
}

func (a *Agent) Run(ctx context.Context, input *tools.AgentInput) *tools.AgentOutput {
	a.sessionId = time.Now().Format("20060102150405")
	a.timeline = &utils.Timeline{}
	a.timeline.Init(a.sessionId)
	a.allowMap = make(map[string]bool)
	a.output = berio.NewCliOutput()
	output := a.output
	a.ignore = utils.NewIgnore(".", []string{".gitignore", ".bergoignore"})
	if config.GlobalConfig.Debug {
		a.agentMode = prompt.MODE_DEBUG
	} else {
		a.agentMode = prompt.MODE_AGENT
	}
	a.waitForUser = true
	a.multiline = false
	a.initCmdHandler()
	a.initToolHandler()
	modelConf := config.GlobalConfig.GetModelConfig(config.GlobalConfig.MainModel)
	if modelConf == nil {
		panic(fmt.Sprintf("main model %s not found", config.GlobalConfig.MainModel))
	}
	a.stats.WindowSize = modelConf.ContextWindow
	for {
		if a.stop {
			break
		}
		output.Stop()
		userInput := a.readFromUser()
		if userInput == "" {
			continue
		}
		a.InteruptNum = 0
		filtered, goToStart := a.handleCmd(userInput)
		if goToStart {
			continue
		}
		filtered, ok := a.processAtCommand(filtered)
		if !ok {
			continue
		}
		a.timeline.InitCheckpoint()
		query := utils.Query{}
		query.SetUserInput(filtered)
		query.SetAttachment(a.attachments)
		query.SetMode(a.agentMode)
		query.SetMememtoNotice()
		a.attachments = nil
		utils.InitMementoFile(a.sessionId)
		a.saveCheckPoint()
		output.OnSystemMsg(utils.UserQueryStyle(filtered), berio.MsgTypeDump)
		utils.AddSessionItem(a.sessionId, filtered)
		if !a.timeline.CanAddQuery() {
			a.timeline.ReplaceLastUserInput(&query)
		} else {
			a.timeline.AddUserInput(&query)
		}
		cli.PrintDebugText("query: \n%v", query.Build())
		a.doTask(ctx)
	}
	return &tools.AgentOutput{}
}
func (a *Agent) captureSignal(cancel context.CancelFunc) {
	//capture signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		cancel()
		signal.Reset(os.Interrupt)
	}()
}
func StopGap() {
	var input string
	fmt.Scanf("%s", &input)
}

func isChanClose(signalChan chan os.Signal) bool {
	select {
	case _, ok := <-signalChan:
		return !ok
	default:
		return false
	}
}
func (a *Agent) doTask(ctx context.Context) {
	defer utils.HideMementoFile(a.sessionId)
	//clean tail tool calls
	a.timeline.CleanTailToolCalls()
	a.timeline.SetTaskEpoch()
	mainModelConf := config.GlobalConfig.GetModelConfig(config.GlobalConfig.MainModel)
	output := a.output
	toolCallAnswers := []*tools.AgentOutput{}
	keepGoing := true
	signalChan := make(chan os.Signal, 1)
	defer func() {
		if !isChanClose(signalChan) {
			close(signalChan)
		}
		signal.Reset(os.Interrupt)
	}()
	for {
		if !isChanClose(signalChan) {
			close(signalChan)
		}
		signal.Reset(os.Interrupt)
		output.Stop() //just in case
		if a.stop {
			break
		}
		if len(toolCallAnswers) == 0 && !keepGoing {
			break
		}
		keepGoing = false
		chatItems := a.timeline.GetChatContext(true)
		if cli.Debug {
			reasoningItem := 0
			for _, item := range chatItems {
				if item.Role == "assistant" && len(item.ReasoningContent) > 0 {
					reasoningItem++
				}
			}
			cli.PrintDebugText("len of chatItems: %d, reasoningItem: %d", len(chatItems), reasoningItem)
		}
		chatItems = llm.InjectSystemPrompt(chatItems, prompt.GetSystemPrompt())
		ctxWithCancel, cancel := context.WithCancel(context.Background())
		cli.SetCancelFunc(cancel)

		signalChan = make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		go func() {
			<-signalChan
			cancel()
		}()
		content := bytes.NewBuffer(nil)
		reasoningContent := bytes.NewBuffer(nil)

		streamer, err := utils.NewLlmStreamer(ctxWithCancel, mainModelConf, chatItems, a.toolSchema)
		if err != nil {
			output.OnSystemMsg(locales.Sprintf("error: %v", err), berio.MsgTypeWarning)
			break
		}
		output.OnSystemMsg(utils.LLMInputStyle("🤖|Bergo: "), berio.MsgTypeDump)
		for streamer.Next() {
			rc, c, toolNames := streamer.ReadWithTool()
			content.WriteString(c)
			reasoningContent.WriteString(rc)
			if len(rc) > 0 {
				output.OnLLMResponse(rc, true)
			}
			if len(c) > 0 {
				output.OnLLMResponse(c, false)
			}
			for _, toolName := range toolNames {
				desc := tools.ToolsMap[toolName]
				if desc != nil && desc.Intent != "" {
					a.output.UpdateTail(utils.ToolUseStyle(desc.Intent))
				}
			}
		}
		if streamer.Error() != nil {
			output.OnSystemMsg(locales.Sprintf("error: %v", streamer.Error()), berio.MsgTypeWarning)
			break
		}

		renderedContent := output.Stop()

		toolCallRequests := streamer.ToolCalls()
		toolCallAnswers = nil
		rc := reasoningContent.String()

		if streamer.TokenUsage.TotalTokens > 0 {
			a.stats.SetTokenUsage(&streamer.TokenUsage)
		}
		//超过窗口，进行压缩
		if a.stats.WindowSize != 0 && float64(a.stats.TokenUsageSession.TotalTokens) > float64(a.stats.WindowSize)*config.GlobalConfig.CompactThreshold {
			a.compact(ctxWithCancel)
			keepGoing = true
			continue
		}
		a.timeline.AddLLMResponse(content.String(), rc, renderedContent, toolCallRequests, streamer.Signature())

		//Tool use if any
		hasStopLoop := false
		if config.GlobalConfig.Debug {
			var tools []string
			for _, call := range toolCallRequests {
				tools = append(tools, call.Function.Name)
			}
			cli.PrintDebugText("tool calls: %v", tools)
		}
		if len(toolCallRequests) > 0 {
			for _, call := range toolCallRequests {
				if call.Function.Name == tools.TOOL_STOP_LOOP {
					hasStopLoop = true
				}
				cli.PrintDebugText("calling tool: %v", call.Function.Name)
				answer, err := a.doToolUse(ctxWithCancel, call)
				if err != nil {
					a.output.OnSystemMsg(locales.Sprintf("error when tool call: %v", err), berio.MsgTypeWarning)
					return
				}
				if answer == nil {
					continue
				}
				toolCallAnswers = append(toolCallAnswers, answer)
			}
		}
		if hasStopLoop {
			break
		}
		if len(toolCallAnswers) > 0 {
			for _, answer := range toolCallAnswers {
				a.timeline.AddToolCallResult(answer.ToolCall.ID, answer.ToolCall.Function.Name, answer.Content, answer.Rendered)
				cli.PrintDebugText("%s\n%s\n%s\n", answer.ToolCall.ID, answer.ToolCall.Function.Name, answer.Content)
			}
		}
	}

}
func (a *Agent) getCliInput() berio.BerInput {
	attachments := make([]string, len(a.attachments))
	for i, attachment := range a.attachments {
		attachments[i] = filepath.Base(attachment.Path)
	}
	return berio.NewCliInput(cli.InputOptions{
		Mode:        a.agentMode,
		Attachments: attachments,
		Stats:       a.stats,
		Multiline:   a.multiline,
		Model:       config.GlobalConfig.MainModel,
		SessionId:   a.timeline.SessionId,
	})
}
func (a *Agent) readFromUser() string {
	receiver := a.getCliInput()
	if a.multiline {
		a.multiline = false
	}
	content, err := receiver.Read()
	if err == cli.ErrInterrupt {
		cli.PrintDebugText("%v", err.Error())
		a.InteruptNum++
		if a.InteruptNum == 1 {
			a.output.OnSystemMsg(locales.Sprintf("Press Ctrl+C or ESC again to exit"), berio.MsgTypeText)
			return ""
		} else if a.InteruptNum == 2 {
			a.stop = true
			return ""
		}
		return ""
	}
	content = strings.TrimSpace(content)
	return content
}

func (a *Agent) saveCheckPoint() {
	commit := "auto save"
	hash := a.timeline.CheckpointSave(commit, a.stats.TokenUsageSession)
	if hash != "" {
		a.output.OnSystemMsg(locales.Sprintf("checkpoint saved, hash: %v", hash), berio.MsgTypeText)
	}
}

func (a *Agent) doToolUse(ctx context.Context, call *llm.ToolCall) (*tools.AgentOutput, error) {
	if handler, ok := a.toolHandler[call.Function.Name]; ok {
		desc := tools.ToolsMap[call.Function.Name]
		err := tools.JsonSchemaExam(call)
		if err != nil {
			a.output.OnSystemMsg(locales.Sprintf("error when calling [%s] err: %v", call.Function.Name, err), berio.MsgTypeWarning)
			return &tools.AgentOutput{
				Content:  err.Error(),
				ToolCall: call,
			}, nil
		}
		chats := a.timeline.GetChatContext(false)
		input := &tools.AgentInput{
			ToolCall:  call,
			Output:    a.output,
			Ig:        a.ignore,
			Timeline:  a.timeline,
			Input:     a.getCliInput(),
			AllowMap:  a.allowMap,
			TaskChats: chats,
		}
		answer := handler(ctx, input)
		if answer.ToolCall == nil {
			answer.ToolCall = call
		}
		if answer.InterruptErr != nil {
			return nil, answer.InterruptErr
		}
		if answer.Error != nil {
			a.output.OnSystemMsg(locales.Sprintf("error when calling [%s] err: %v", call.Function.Name, answer.Error), berio.MsgTypeWarning)
			return &tools.AgentOutput{
				Content:  answer.Error.Error(),
				ToolCall: call,
			}, nil
		}
		rendered := ""
		if desc.OutputFunc != nil {
			rendered = desc.OutputFunc(call, answer.Content)
			a.output.OnSystemMsg(rendered, berio.MsgTypeDump)
		}
		return &tools.AgentOutput{
			Content:  answer.Content,
			ToolCall: answer.ToolCall,
			Rendered: rendered,
		}, nil

	}
	return nil, nil
}

func (a *Agent) compact(ctx context.Context) {
	chats := a.timeline.GetChatContext(false)
	out := tools.Compact(ctx, &tools.AgentInput{
		TaskChats: chats,
		Output:    a.output,
	})
	if out.Error != nil {
		a.output.OnSystemMsg(fmt.Sprintf("compact error %v", out.Error), berio.MsgTypeWarning)
		return
	}
	a.timeline.AddCompact()
}

var atCmdRegex = regexp.MustCompile(`@\S+`)

func (a *Agent) processAtCommand(userInput string) (string, bool) {
	matches := atCmdRegex.FindAllString(userInput, -1)
	// 将提取的@字符串转换为Attachment对象并存储
	index := 1
	var attachments []*utils.Attachment
	for _, match := range matches {
		if strings.HasPrefix(match, "@file:") {
			path := strings.TrimPrefix(match, "@file:")
			stat, err := os.Stat(path)
			if err != nil {
				a.output.OnSystemMsg(locales.Sprintf("invalid file path: %v", path), berio.MsgTypeWarning)
				return "", false
			}
			userInput = strings.ReplaceAll(userInput, match, fmt.Sprintf("[bergo-attch %d]", index))
			if stat.IsDir() {
				attachments = append(attachments, &utils.Attachment{
					Index: index,
					Path:  path,
					Type:  utils.AttachmentTypeDir,
				})
			} else {
				attachments = append(attachments, &utils.Attachment{
					Index: index,
					Path:  path,
					Type:  utils.AttachmentTypeFile,
				})
			}
			index++
		}

	}
	a.attachments = attachments
	return userInput, true
}
