package main

import (
	"bergo/agent"
	"bergo/config"
	"bergo/locales"
	"bergo/utils"
	"bergo/utils/cli"
	"bergo/wizard"
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/pterm/pterm"
)

// 使用wizard.ProviderConfig代替硬编码的ProviderInfo

// readInput 读取用户输入（不使用TeaInput）
func readInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// readInputHidden 读取敏感信息（如API密钥），尝试隐藏输入
func readInputHidden(prompt string) (string, error) {
	fmt.Print(prompt)
	
	// 尝试使用系统特定的方式隐藏输入
	// 这里使用简单的读取，在实际使用中可以考虑使用golang.org/x/term
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// showWizardHeader 显示向导标题
func showWizardHeader() {
	fmt.Println("")
	pterm.DefaultHeader.WithFullWidth().Println("Bergo 配置向导")
	fmt.Println("")
}

// confirmOverwrite 确认是否覆盖现有配置文件
func confirmOverwrite(configPath string) bool {
	if _, err := os.Stat(configPath); err == nil {
		pterm.Warning.Println("检测到已存在的配置文件:", configPath)
		
		overwrite, err := pterm.DefaultInteractiveConfirm.Show("是否覆盖现有配置文件?")
		if err != nil {
			pterm.Error.Println("确认操作失败:", err)
			return false
		}
		
		if !overwrite {
			pterm.Info.Println("已取消配置向导")
			return false
		}
	}
	return true
}

// selectProvider 选择模型提供商
func selectProvider() *wizard.ProviderConfig {
	pterm.Info.Println("步骤 1/3: 选择主要模型提供商")
	fmt.Println("请选择您要使用的AI模型提供商:")
	
	providers, err := wizard.GetProviders()
	if err != nil {
		pterm.Error.Println("加载提供商配置失败:", err)
		return nil
	}
	
	providerOptions := make([]string, len(providers))
	for i, p := range providers {
		providerOptions[i] = fmt.Sprintf("%s - %s", p.DisplayName, p.Description)
	}
	
	selectedOption, err := pterm.DefaultInteractiveSelect.
		WithOptions(providerOptions).
		Show()
	
	if err != nil {
		pterm.Error.Println("选择提供商失败:", err)
		return nil
	}
	
	// 找到选中的提供商
	for i, p := range providers {
		option := fmt.Sprintf("%s - %s", p.DisplayName, p.Description)
		if option == selectedOption {
			pterm.Success.Printf("已选择: %s\n\n", p.DisplayName)
			return &providers[i]
		}
	}
	
	pterm.Error.Println("未选择有效的提供商")
	return nil
}

// inputAPIKey 输入API密钥
func inputAPIKey(provider *wizard.ProviderConfig) string {
	pterm.Info.Println("步骤 2/3: 输入API密钥")
	fmt.Printf("请输入您的 %s API密钥:\n", provider.DisplayName)
	
	apiKey, err := readInputHidden(fmt.Sprintf("%s API密钥: ", provider.DisplayName))
	if err != nil {
		pterm.Error.Println("输入错误:", err)
		return ""
	}
	
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		pterm.Error.Println("API密钥不能为空!")
		return ""
	}
	
	pterm.Success.Println("API密钥已保存\n")
	return apiKey
}

// selectModel 选择主模型
func selectModel(provider *wizard.ProviderConfig) string {
	pterm.Info.Println("步骤 3/3: 选择主模型")
	
	selectedOption, err := pterm.DefaultInteractiveSelect.
		WithOptions(provider.Models).
		Show()
	
	if err != nil {
		pterm.Error.Println("选择模型失败:", err)
		return ""
	}
	
	selectedModel := ""
	if selectedOption == "自定义模型" {
		fmt.Println("请输入自定义模型名称:")
		customModel, err := readInput("模型名称 (例如: gpt-4, claude-3-opus): ")
		if err != nil {
			pterm.Error.Println("输入错误:", err)
			return ""
		}
		selectedModel = strings.TrimSpace(customModel)
	} else {
		// 移除推荐标记
		selectedModel = strings.Replace(selectedOption, " (推荐)", "", 1)
	}
	
	if selectedModel == "" {
		selectedModel = provider.DefaultModel
	}
	
	pterm.Success.Printf("已选择模型: %s\n\n", selectedModel)
	return selectedModel
}

// showSuccessMessage 显示成功信息
func showSuccessMessage(configPath string) {
	pterm.Success.Println("配置文件创建成功!")
	fmt.Println("")
	pterm.Info.Println("配置文件路径:", configPath)
	fmt.Println("")
	pterm.Println("您可以使用以下命令启动Bergo:")
	pterm.Println(pterm.LightCyan("  bergo " + configPath))
	fmt.Println("")
	pterm.Println("或者将配置文件重命名为 bergo.toml 并直接运行:")
	pterm.Println(pterm.LightCyan("  bergo"))
	fmt.Println("")
	pterm.Info.Println("祝您使用愉快!")
}

// RunInitWizard 运行初始化配置向导
func RunInitWizard() {
	showWizardHeader()
	
	configPath := "bergo.toml"
	if !confirmOverwrite(configPath) {
		return
	}
	
	provider := selectProvider()
	if provider == nil {
		return
	}
	
	apiKey := inputAPIKey(provider)
	if apiKey == "" {
		return
	}
	
	model := selectModel(provider)
	if model == "" {
		return
	}
	
	config := createConfig(provider, apiKey, model)
	if err := saveConfig(configPath, config); err != nil {
		pterm.Error.Println("保存配置文件失败:", err)
		return
	}
	
	showSuccessMessage(configPath)
}

// createConfig 创建配置对象
func createConfig(provider *wizard.ProviderConfig, apiKey string, mainModel string) map[string]interface{} {
	config := make(map[string]interface{})

	// 基本配置
	config["debug"] = false
	config["language"] = "chinese"
	config["line_budget"] = 1000
	config["compact_threshold"] = 0.8

	// 设置主模型
	config["main_model"] = mainModel
	
	// 根据主模型类型设置berag模型
	beragModel, err := wizard.GetNonReasoningModel(mainModel)
	if err != nil {
		pterm.Warning.Printf("获取berag模型失败，使用主模型: %v\n", err)
		beragModel = mainModel
	}
	config["berag_model"] = beragModel
	config["berag_extract_model"] = beragModel

	// 设置API密钥
	config[provider.APIKeyField] = apiKey

	return config
}

// saveConfig 保存配置到文件
func saveConfig(path string, config map[string]interface{}) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// 转换为TOML
	tomlBytes, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(path, tomlBytes, 0644)
}

var bergoTitle = `
██████╗ ███████╗██████╗  ██████╗  ██████╗ 
██╔══██╗██╔════╝██╔══██╗██╔═══╗  ██╔═══██╗
██████╔╝█████╗  ██████╔╝██║ ████║██║   ██║       
██╔══██╗██╔══╝  ██╔══██╗██║   ██║██║   ██║
██████╔╝███████╗██║  ██║╚██████╔╝╚██████╔╝
╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝  ╚═════╝ 
`

func readConfig() {
	if len(os.Args) > 1 {
		err := config.ReadConfig(os.Args[1])
		if err != nil {
			panic(err)
		}
	}
	if config.GlobalConfig == nil {
		panic(locales.Sprintf("config is nil"))
	}
}

func main() {
	utils.EnvInit()

	// 检查是否有init命令
	if len(os.Args) > 1 && os.Args[1] == "init" {
		fmt.Println(bergoTitle)
		RunInitWizard()
		return
	}

	readConfig()
	cli.Debug = config.GlobalConfig.Debug
	fmt.Println(bergoTitle)
	mp := agent.NewMainAgent()
	ctx, _ := context.WithCancel(context.Background())
	mp.Run(ctx, nil)
}