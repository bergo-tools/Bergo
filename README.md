# Bergo - Bot Engineer in Golang

![Bergo Logo](./bergo-logo.png)

```
██████╗ ███████╗██████╗  ██████╗  ██████╗ 
██╔══██╗██╔════╝██╔══██╗██╔═══╗  ██╔═══██╗
██████╔╝█████╗  ██████╔╝██║ ████║██║   ██║       
██╔══██╗██╔══╝  ██╔══██╗██║   ██║██║   ██║
██████╔╝███████╗██║  ██║╚██████╔╝╚██████╔╝
╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝  ╚═════╝ 
```

## Bergo

BERGO(Bot EngineeR in GO) 是一个基于 Golang 开发的 Coding Agent，它可以帮助开发者快速、高效地完成代码编写任务。
本项目主要的代码主要由AI编写，质量比较粗糙。

## 特点

- **多模型支持**: 不限定模型，支持多种API提供商
- **开箱即用**: 丰富的配置，开箱即用
- **CheckPoint支持**: 支持代码检查点，随时回退你的代码
- **Agentic RAG**: 内置智能RAG能力
- **多模式支持**: 支持ASK、PLANNER、AGENT等多种工作模式
- **丰富工具集**: 提供文件编辑、代码分析、Shell命令等多种工具
- **多语言支持**: 完整的国际化支持

## 预置模型

### LLM提供商
- **OpenAI**: GPT系列模型
- **Anthropic**: Claude系列模型  
- **DeepSeek**: DeepSeek系列模型
- **Kimi**: Kimi系列模型
- **MiniMax**: MiniMax系列模型
- **OpenRouter**: 支持多种第三方模型
- **Mock**: 用于测试的模拟模型

### 配置文件示例

项目使用 `bergo.toml` 配置文件：

```toml
# 主模型
main_model = "minimax-m2"
# berag模型
berag_model = "minimax-m2"
# berag提取模型
berag_extract_model = "minimax-m2"
# 读取最大行数
line_budget = 1000
# 调试模式
debug = false

# API密钥配置，开箱即用支持3种模型供应商，deepseek/kimi/minimax
deepseek_api_key = "your-deepseek-key"
kimi_api_key = "your-kimi-key"
minimax_api_key = "your-minimax-key"

# 模型配置
[[models]]
identifier = "mock"
provider = "mock"
reasoning_tag = "think"


```

## 预置命令

### 系统命令
- `/exit` - 退出程序
- `/help` - 显示帮助信息
- `/clear` - 清除当前会话，创建新会话
- `/sessions` - 加载历史会话

### 模式切换
- `/ask` - 切换到ASK模式
- `/planner` - 切换到PLANNER模式  
- `/agent` - 切换到AGENT模式
- `/multiline` - 启用多行输入模式

### 功能命令
- `/timeline` - 查看操作时间线
- `/revert` - 回退到历史状态
- `/model` - 切换模型
- `/compact` - 压缩上下文

## 核心模块

### Agent模块 (`agent/`)
- `agent.go`: 主要智能体逻辑
- `cmd.go`: 命令处理系统
- `tools.go`: 工具管理

### 工具模块 (`tools/`)
- `edit.go`: 文件编辑工具
- `read_file.go`: 文件读取工具
- `shell_cmd.go`: Shell命令执行工具
- `berag.go`: RAG增强工具
- `task.go`: 任务管理工具
- `remove.go`: 文件删除工具
- `stop_loop.go`: 循环控制工具
- `compact.go`: 上下文压缩工具

### LLM模块 (`llm/`)
- `provider.go`: LLM提供商接口
- `openai.go`: OpenAI提供商
- `anthropic.go`: Anthropic提供商
- `deepseek.go`: DeepSeek提供商
- `kimi.go`: Kimi提供商
- `minimax.go`: MiniMax提供商
- `openrouter.go`: OpenRouter提供商
- `mock.go`: 模拟提供商

### 其他模块
- `berio/`: 输入输出处理
- `locales/`: 国际化支持
- `config/`: 配置管理
- `utils/`: 工具函数库
- `prompt/`: 提示词管理

## 使用方法

### 安装依赖
```bash
go mod download
```

### 运行程序
```bash
# 使用默认配置
go run .

# 指定配置文件
go run . bergo.toml
```


### Memento文件

Bergo会在工作目录下维护一个 `.bergo.memento` 文件，用于在长时间对话中保持上下文记忆, 你可能会在Bergo工作时短暂看到它


## 开发指南

### 项目结构
```
bergo/
├── agent/          # 智能体核心逻辑
├── berio/          # 输入输出处理
├── config/         # 配置管理
├── llm/            # LLM提供商
├── locales/        # 国际化
├── prompt/         # 提示词管理
├── tools/          # 工具集
├── utils/          # 工具函数
├── bergo.toml      # 配置文件
├── main.go         # 主程序入口
└── README.md       # 项目说明
```

### 添加新的LLM提供商

1. 在 `llm/` 目录下创建新的提供商文件
2. 实现 `LLMProvider` 接口
3. 在配置文件中添加相应配置

### 自定义工具

1. 在 `tools/` 目录下创建新工具文件
2. 实现 `AgentInput` 和 `AgentOutput` 接口
3. 在 `agent.go` 中注册新工具

## 技术栈

- **Go 1.25**: 主要开发语言
- **Bubble Tea**: 现代终端UI框架


## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进本项目。

---

**Bergo** - 本文件由AI编写 🚀