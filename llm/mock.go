package llm

import (
	"bergo/config"
	"context"
	"fmt"
	"time"
)

type MockProvider struct {
	Identifier string
}

var NormalMockResponses [][]*Response
var normalIdx int
var BeragMockResponses [][]*Response
var beragIdx int
var BeragExtractResponses [][]*Response
var beragExtractIdx int

func init() {
	NormalMockResponses = [][]*Response{
		{
			{
				Content: "This is a mock response 2",
			},
			{
				Content: "This is a mock response 3",
			},
			{
				Content: "<todo_list><create>item 1\n",
			},
			{
				Content: "item 2\n</create></todo_list>",
				ToolCalls: []*ToolCall{
					{
						ID:   "xxx1",
						Type: "function",
						Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{
							Name:      "shell_cmd",
							Arguments: "{\"command\":\"ls -l\"}",
						},
					},
				},
			},
			{
				TokenStatics: &TokenUsage{
					TotalTokens:      100,
					PromptTokens:     50,
					CompletionTokens: 50,
					CachedTokens:     25,
				},
				FinishReason: FinishReasonStop,
			},
		},
	}

	BeragExtractResponses = append(BeragExtractResponses, []*Response{
		{
			Content: "try to extract",
		},
		{
			Content: "<extract_result><extract_item>main.go\n</extract_item></extract_result>",
		},
	})
	BeragMockResponses = append(BeragMockResponses, []*Response{
		{
			Content: "begin task\n",
		},
		{
			Content: "<shell_cmd>ls</shell_cmd>\n",
		},
		{
			Content: "<shell_cmd>ls -l</shell_cmd>\n",
		},
		{
			Content: "<shell_cmd>git status</shell_cmd>\n",
		},
		{
			Content: "<berag_extract>main.go</berag_extract>\n",
		},
		{
			Content: "<berag_extract>main.go</berag_extract>\n",
		},
		{
			Content: "<berag_extract>main.go</berag_extract>\n",
		},
	})
	BeragMockResponses = append(BeragMockResponses, []*Response{
		{
			Content: "<stop_loop></stop_loop>\n",
		},
	})
}

func (p *MockProvider) Init(conf *config.ModelConfig) error {
	p.Identifier = conf.Identifier
	return nil
}
func (p *MockProvider) StreamResponse(ctx context.Context, req *Request) <-chan *Response {
	var responses [][]*Response
	var idx *int
	switch p.Identifier {
	case "berag":
		responses = BeragMockResponses
		idx = &beragIdx
	case "extract":
		responses = BeragExtractResponses
		idx = &beragExtractIdx
	case "mock":
		responses = NormalMockResponses
		idx = &normalIdx
	default:
		panic(fmt.Sprintf("mock provider %s not found", p.Identifier))
	}

	ch := make(chan *Response)
	go func() {
		time.Sleep(2 * time.Second)
		defer close(ch)
		for _, resp := range responses[*idx] {
			ch <- resp
			time.Sleep(2 * time.Second)
		}
		(*idx)++
		if *idx >= len(responses) {
			*idx = 0
		}
	}()
	return ch
}
func (p *MockProvider) StreamResponseWithImgInput(ctx context.Context, req *Request) <-chan *Response {
	return p.StreamResponse(ctx, req)
}
func (p *MockProvider) ListModels() ([]string, error) {
	return nil, nil
}
func (p *MockProvider) Close() error {
	return nil
}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}
