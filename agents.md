# 项目纵览
这是一个基于Go语言的项目，是一个命令行的AI Coding Agent。它连接各种LLM的api，并解决用户提出的问题

# 目录结构
- /agent agent的主流程代码
- /config 配置加载相关的逻辑
- /prompt 包含system prompt模版
- /tools agent可以调用的Tools
- /utils 工具函数，很多实际执行的逻辑都是写在这里面的
- /utils/cli 命令行tui相关的逻辑
- /llm llm api调用相关的逻辑
- /locales 国际化相关
- /test 测试相关的代码
- main.go 项目的入口文件

# 注意事项
1. 输出字符串考虑使用locales.Sprintf函数，方便国际化，locales内有提取locales.Sprintf中字符串的工具函数，用来做国际化
2. 测试尽量放/test目录下