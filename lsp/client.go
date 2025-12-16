package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LSP消息结构
type Message struct {
	Jsonrpc string         `json:"jsonrpc"`
	ID      interface{}    `json:"id,omitempty"`
	Method  string         `json:"method,omitempty"`
	Params  interface{}    `json:"params,omitempty"`
	Result  interface{}    `json:"result,omitempty"`
	Error   *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
type LspClient interface {
	Initialize(ctx context.Context, rootPath string) error
	Shutdown(ctx context.Context) error
	FindReference(ctx context.Context, filePath string, line int, character int) ([]Location, error)
	FindDefinition(ctx context.Context, filePath string, line int, character int) ([]Location, error)
}

// 基础LSP客户端
type BaseClient struct {
	stdin            io.WriteCloser
	stdout           io.ReadCloser
	serverProc       *os.Process
	rootPath         string
	workspaceFolders []string // 支持多工作区
	mu               sync.Mutex
	requestID        int
}

// 客户端配置
type ClientConfig struct {
	ServerCommand    []string
	RootPath         string
	WorkspaceFolders []string // 支持多工作区
}

// LSP位置相关结构
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// 工作区文件夹
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// 初始化参数
type InitializeParams struct {
	ProcessID        int                `json:"processId"`
	RootPath         string             `json:"rootPath,omitempty"`
	RootURI          string             `json:"rootUri,omitempty"`
	Capabilities     ClientCapabilities `json:"capabilities"`
	WorkspaceFolders []WorkspaceFolder  `json:"workspaceFolders,omitempty"`
}

type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Workspace    WorkspaceClientCapabilities    `json:"workspace,omitempty"`
}

type TextDocumentClientCapabilities struct {
	Definition *TextDocumentDefinitionCapabilities `json:"definition,omitempty"`
	References *TextDocumentReferencesCapabilities `json:"references,omitempty"`
}

type TextDocumentDefinitionCapabilities struct {
	LinkSupport bool `json:"linkSupport,omitempty"`
}

type TextDocumentReferencesCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type WorkspaceClientCapabilities struct {
	WorkspaceFolders bool `json:"workspaceFolders,omitempty"`
}

// 初始化结果
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

type ServerCapabilities struct {
	DefinitionProvider bool        `json:"definitionProvider,omitempty"`
	ReferencesProvider bool        `json:"referencesProvider,omitempty"`
	TextDocumentSync   interface{} `json:"textDocumentSync,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// 文本文档位置参数
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// 引用参数
type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// 创建新的LSP客户端
func NewLspClient(config ClientConfig) (*BaseClient, error) {
	if len(config.ServerCommand) == 0 {
		return nil, fmt.Errorf("server command is required")
	}

	// 启动LSP服务器进程
	cmd := exec.Command(config.ServerCommand[0], config.ServerCommand[1:]...)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start LSP server: %w", err)
	}
	time.Sleep(100 * time.Millisecond)

	client := &BaseClient{
		stdin:            stdin,
		stdout:           stdout,
		serverProc:       cmd.Process,
		rootPath:         config.RootPath,
		workspaceFolders: config.WorkspaceFolders,
	}

	return client, nil
}

// 初始化LSP服务器
func (c *BaseClient) Initialize(ctx context.Context, rootPath string) error {
	if rootPath == "" {
		rootPath = c.rootPath
	}

	// 获取工作目录的绝对路径
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 构建工作区文件夹列表
	var workspaceFolders []WorkspaceFolder = nil

	// 如果有配置的工作区文件夹，使用它们
	if len(c.workspaceFolders) > 0 {
		for _, folder := range c.workspaceFolders {
			absFolderPath, err := filepath.Abs(folder)
			if err != nil {
				log.Printf("Warning: failed to get absolute path for workspace folder %s: %v", folder, err)
				continue
			}
			workspaceFolders = append(workspaceFolders, WorkspaceFolder{
				URI:  "file://" + absFolderPath,
				Name: filepath.Base(absFolderPath),
			})
		}
	} else {
		// 否则使用单个根路径
		workspaceFolders = append(workspaceFolders, WorkspaceFolder{
			URI:  "file://" + absRootPath,
			Name: filepath.Base(absRootPath),
		})
	}

	params := InitializeParams{
		ProcessID: int(os.Getpid()),
		RootPath:  absRootPath,
		RootURI:   "file://" + absRootPath,
		Capabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				Definition: &TextDocumentDefinitionCapabilities{
					LinkSupport: true,
				},
				References: &TextDocumentReferencesCapabilities{
					DynamicRegistration: true,
				},
			},
			Workspace: WorkspaceClientCapabilities{
				WorkspaceFolders: true,
			},
		},
		WorkspaceFolders: workspaceFolders,
	}

	var result InitializeResult
	if err := c.sendRequest(ctx, "initialize", params, &result); err != nil {
		// 尝试获取更详细的错误信息
		//log.Printf("Initialize request details - Method: initialize, Params: %+v", params)
		return fmt.Errorf("initialize failed: %w", err)
	}

	//log.Printf("LSP server capabilities: %+v", result.Capabilities)

	// 发送initialized通知
	if err := c.sendNotification(ctx, "initialized", map[string]interface{}{}); err != nil {
		return fmt.Errorf("initialized notification failed: %w", err)
	}

	//log.Printf("LSP server initialized: %s", result.ServerInfo.Name)
	return nil
}

// 查找定义
func (c *BaseClient) FindDefinition(ctx context.Context, filePath string, line int, character int) ([]Location, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	params := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{
			URI: "file://" + absPath,
		},
		Position: Position{
			Line:      line,
			Character: character,
		},
	}

	var result []Location
	if err := c.sendRequest(ctx, "textDocument/definition", params, &result); err != nil {
		return nil, fmt.Errorf("definition request failed: %w", err)
	}

	return result, nil
}

// 查找引用
func (c *BaseClient) FindReference(ctx context.Context, filePath string, line int, character int) ([]Location, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	params := ReferenceParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file://" + absPath,
			},
			Position: Position{
				Line:      line,
				Character: character,
			},
		},
		Context: ReferenceContext{
			IncludeDeclaration: true,
		},
	}

	var result []Location
	if err := c.sendRequest(ctx, "textDocument/references", params, &result); err != nil {
		return nil, fmt.Errorf("references request failed: %w", err)
	}

	return result, nil
}

// 关闭连接
func (c *BaseClient) Shutdown(ctx context.Context) error {
	if err := c.sendRequest(ctx, "shutdown", nil, nil); err != nil {
		log.Printf("Shutdown request failed: %v", err)
	}

	if err := c.sendNotification(ctx, "exit", nil); err != nil {
		log.Printf("Exit notification failed: %v", err)
	}

	if c.serverProc != nil {
		if err := c.serverProc.Kill(); err != nil {
			log.Printf("Failed to kill server process: %v", err)
		}
	}

	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}

	return nil
}

// 发送请求
func (c *BaseClient) sendRequest(ctx context.Context, method string, params interface{}, result interface{}) error {
	c.mu.Lock()
	c.requestID++
	id := c.requestID
	c.mu.Unlock()

	msg := Message{
		Jsonrpc: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := c.writeMessage(msg); err != nil {
		return err
	}

	// 读取响应
	response, err := c.readResponse(id)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("LSP error: %s (code: %d)", response.Error.Message, response.Error.Code)
	}

	if result != nil && response.Result != nil {
		data, err := json.Marshal(response.Result)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, result)
	}

	return nil
}

// 发送通知
func (c *BaseClient) sendNotification(ctx context.Context, method string, params interface{}) error {
	msg := Message{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
	}

	return c.writeMessage(msg)
}

// 写入消息
func (c *BaseClient) writeMessage(msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	fullData := append([]byte(header), data...)

	_, err = c.stdin.Write(fullData)
	return err
}

// 读取响应
func (c *BaseClient) readResponse(expectedID int) (*Message, error) {
	for {
		msg, err := c.readMessage()
		if err != nil {
			return nil, err
		}

		// 只处理对应ID的响应
		if msg.ID != nil {
			if id, ok := msg.ID.(float64); ok && int(id) == expectedID {
				return msg, nil
			}
		}
	}
}

// 读取消息
func (c *BaseClient) readMessage() (*Message, error) {
	reader := bufio.NewReader(c.stdout)

	// 读取头部
	var contentLength int
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("connection closed by server")
			}
			return nil, fmt.Errorf("failed to read header line: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		if strings.HasPrefix(line, "Content-Length: ") {
			lengthStr := strings.TrimPrefix(line, "Content-Length: ")
			if _, err := fmt.Sscanf(lengthStr, "%d", &contentLength); err != nil {
				return nil, fmt.Errorf("failed to parse content length: %w", err)
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("no content length found")
	}

	// 读取内容
	content := make([]byte, contentLength)
	_, err := io.ReadFull(reader, content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	//log.Printf("Received raw LSP message: %s", string(content))

	var msg Message
	if err := json.Unmarshal(content, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w, content: %s", err, string(content))
	}

	return &msg, nil
}

var mapLangToClient = map[string]func() (LspClient, error){
	"go":     NewGoLspClient,
	"golang": NewGoLspClient,
}

func GetClientFunc(lang string) (func() (LspClient, error), bool) {
	clientFunc, ok := mapLangToClient[lang]
	return clientFunc, ok
}
