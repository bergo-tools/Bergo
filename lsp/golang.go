package lsp

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type GoLspClient struct {
	*BaseClient
}

func NewGoLspClient() (LspClient, error) {
	// 检测gopls是否存在
	if _, err := exec.LookPath("gopls"); err != nil {
		return nil, fmt.Errorf("gopls not found in PATH, please install gopls: %w", err)
	}

	config := ClientConfig{
		ServerCommand: []string{"gopls", "serve"},
		RootPath:      ".",
	}
	// 创建基础LSP客户端
	baseClient, err := NewLspClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create base LSP client: %w", err)
	}

	// 创建Go语言特定的LSP客户端
	goClient := &GoLspClient{
		BaseClient: baseClient,
	}

	// 初始化LSP服务器连接
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := goClient.Initialize(ctx, config.RootPath); err != nil {
		// 如果初始化失败，确保清理资源
		goClient.Shutdown(context.Background())
		return nil, fmt.Errorf("failed to initialize LSP server: %w", err)
	}

	return goClient, nil
}
