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
# 语言设置（可选：chinese/english，默认为chinese）
language = "chinese"

# API密钥配置，开箱即用支持3种模型供应商，deepseek/kimi/minimax
deepseek_api_key = "your-deepseek-key"
kimi_api_key = "your-kimi-key"
minimax_api_key = "your-minimax-key"

# 模型配置
[[models]]
identifier = "mock"
provider = "mock"


```

## 快速开始

### 使用初始化向导

Bergo提供了一个简单的初始化向导，帮助您快速创建配置文件：

```bash
# 运行初始化向导
bergo init
```

向导会引导您完成以下步骤：
1. 选择AI模型提供商（DeepSeek、MiniMax、Kimi、Xiaomi、OpenAI、OpenRouter）
2. 输入API密钥
3. 选择主模型
4. 自动创建配置文件


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

### 使用@符号添加上下文

Bergo支持使用`@`符号来添加上下文文件，这在与AI对话时非常有用，可以将相关代码文件作为上下文提供给AI助手。

#### 基本语法
- `@file:path/to/file` - 添加文件或目录作为上下文
- 输入`@file:`后可以触发自动补全，显示当前目录下的文件和文件夹

#### 使用示例
```
# 添加单个文件作为上下文
请帮我分析这个文件 @file:main.go

# 添加整个目录作为上下文
请帮我重构这个模块 @file:utils/

```

#### 注意事项
- 添加的文件内容会被包含在对话上下文中，帮助AI更好地理解代码结构
- 支持添加多个文件，只需在对话中多次使用`@file:`语法
- 文件路径中不能包含空格


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

## 多语言设置

Bergo支持多语言界面，目前支持中文和英文。并且您可以指定LLM回复偏好的语言

### . 界面语言配置
您可以通过环境变量设置语言：
```bash
# 设置环境变量（仅限当前会话）
export BERGO_LANG=zh

# 或者直接运行
BERGO_LANG=zh go run .
```

### llm回复语言偏好
在配置文件中，您可以设置 `language` 字段来指定AI助手的回复语言。

### 默认行为
- 如果未设置LLM默认语言，默认使用中文
- 默认界面语言是英文

## 快速开始

### 使用初始化向导

Bergo提供了一个简单的初始化向导，帮助您快速创建配置文件：

```bash
# 运行初始化向导
bergo init
```

向导会引导您完成以下步骤：
1. 选择AI模型提供商（DeepSeek、MiniMax、Kimi、Xiaomi、OpenAI、OpenRouter）
2. 输入API密钥
3. 选择主模型
4. 自动创建配置文件


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

### 使用@符号添加上下文

Bergo支持使用`@`符号来添加上下文文件，这在与AI对话时非常有用，可以将相关代码文件作为上下文提供给AI助手。

#### 基本语法
- `@file:path/to/file` - 添加文件或目录作为上下文
- 输入`@file:`后按Tab键可以触发自动补全，显示当前目录下的文件和文件夹

#### 使用示例
```
# 添加单个文件作为上下文
请帮我分析这个文件 @file:main.go

# 添加整个目录作为上下文
请帮我重构这个模块 @file:utils/

# 添加相对路径文件
请查看这个配置文件 @file:./config/config.go
```

#### 自动补全功能
1. 输入`@file:`后，按Tab键会显示当前目录下的文件和文件夹列表
2. 继续输入部分路径名，按Tab键会进行智能补全
3. 支持相对路径和绝对路径

#### 注意事项
- 添加的文件内容会被包含在对话上下文中，帮助AI更好地理解代码结构
- 支持添加多个文件，只需在对话中多次使用`@file:`语法
- 文件路径中不能包含空格


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
