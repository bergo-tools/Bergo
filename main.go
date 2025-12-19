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

	"bergo/version"

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
	pterm.DefaultHeader.WithFullWidth().Println(locales.Sprintf("Bergo Configuration Wizard"))
	fmt.Println("")
}

// confirmOverwrite 确认是否覆盖现有配置文件
func confirmOverwrite(configPath string) bool {
	if _, err := os.Stat(configPath); err == nil {
		pterm.Warning.Println(locales.Sprintf("Detected existing configuration file:"), configPath)

		overwrite, err := pterm.DefaultInteractiveConfirm.Show(locales.Sprintf("Overwrite existing configuration file?"))
		if err != nil {
			pterm.Error.Println(locales.Sprintf("Confirmation failed:"), err)
			return false
		}

		if !overwrite {
			pterm.Info.Println(locales.Sprintf("Configuration wizard cancelled"))
			return false
		}
	}
	return true
}

// selectProvider 选择模型提供商
func selectProvider() *wizard.ProviderConfig {
	pterm.Info.Println(locales.Sprintf("Step 1/3: Select primary model provider"))
	fmt.Println(locales.Sprintf("Please select the AI model provider you want to use:"))

	providers, err := wizard.GetProviders()
	if err != nil {
		pterm.Error.Println(locales.Sprintf("Failed to load provider configuration:"), err)
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
		pterm.Error.Println(locales.Sprintf("Failed to select provider:"), err)
		return nil
	}

	// 找到选中的提供商
	for i, p := range providers {
		option := fmt.Sprintf("%s - %s", p.DisplayName, p.Description)
		if option == selectedOption {
			pterm.Success.Print(locales.Sprintf("Selected: %s\n\n", p.DisplayName))
			return &providers[i]
		}
	}

	pterm.Error.Println(locales.Sprintf("No valid provider selected"))
	return nil
}

// inputAPIKey 输入API密钥
func inputAPIKey(provider *wizard.ProviderConfig) string {
	pterm.Info.Println(locales.Sprintf("Step 2/3: Enter API key"))
	fmt.Print(locales.Sprintf("Please enter your %s API key:\n", provider.DisplayName))

	apiKey, err := readInputHidden(fmt.Sprint(locales.Sprintf("%s API key: ", provider.DisplayName)))
	if err != nil {
		pterm.Error.Println(locales.Sprintf("Input error:"), err)
		return ""
	}

	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		pterm.Error.Println(locales.Sprintf("API key cannot be empty!"))
		return ""
	}

	pterm.Success.Println(locales.Sprintf("API key saved\n"))
	return apiKey
}

// selectModel 选择主模型
func selectModel(provider *wizard.ProviderConfig) string {
	pterm.Info.Println(locales.Sprintf("Step 3/3: Select main model"))

	selectedOption, err := pterm.DefaultInteractiveSelect.
		WithOptions(provider.Models).
		Show()

	if err != nil {
		pterm.Error.Println(locales.Sprintf("Failed to select model:"), err)
		return ""
	}

	selectedModel := ""
	if selectedOption == "自定义模型" {
		fmt.Println(locales.Sprintf("Please enter custom model name:"))
		customModel, err := readInput(locales.Sprintf("Model name (e.g., gpt-4, claude-3-opus): "))
		if err != nil {
			pterm.Error.Println(locales.Sprintf("Input error:"), err)
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

	pterm.Success.Print(locales.Sprintf("Selected model: %s\n\n", selectedModel))
	return selectedModel
}

// showSuccessMessage 显示成功信息
func showSuccessMessage(configPath string) {
	pterm.Success.Println(locales.Sprintf("Configuration file created successfully!"))
	fmt.Println("")
	pterm.Info.Println(locales.Sprintf("Configuration file path:"), configPath)
	fmt.Println("")
	pterm.Println(locales.Sprintf("You can start Bergo with the following command:"))
	pterm.Println(pterm.LightCyan("  bergo " + configPath))
	fmt.Println("")
	pterm.Println(locales.Sprintf("Or rename the configuration file to bergo.toml and run directly:"))
	pterm.Println(pterm.LightCyan("  bergo"))
	fmt.Println("")
	pterm.Info.Println(locales.Sprintf("Enjoy using Bergo!"))
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
		pterm.Error.Println(locales.Sprintf("Failed to save configuration file:"), err)
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
		pterm.Warning.Print(locales.Sprintf("Failed to get berag model, using main model: %v\n", err))
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
		RunInitWizard()
		return
	}
	fmt.Println(bergoTitle)
	// 显示完整的版本信息
	versionInfo := version.FormatVersion(version.Version, version.BuildTime, version.CommitHash)
	pterm.Info.Println(fmt.Sprintf("Version %s", versionInfo))
	
	// 检查更新
	version.CheckAndHandleUpdates()
	
	readConfig()

	cli.Debug = config.GlobalConfig.Debug
	mp := agent.NewMainAgent()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mp.Run(ctx, nil)
}
