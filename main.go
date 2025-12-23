package main

import (
	"bergo/agent"
	"bergo/config"
	"bergo/locales"
	"bergo/utils"
	"bergo/utils/cli"
	"bergo/wizard"
	"context"
	"fmt"
	"os"

	"bergo/version"

	"github.com/pterm/pterm"
)

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

// checkAndCleanSessions 检查session数量，如果超过配置的最大值则提示是否清空
func checkAndCleanSessions() {
	maxCount := config.GlobalConfig.MaxSessionCount
	// 如果配置值小于等于0，不进行检查
	if maxCount <= 0 {
		return
	}

	sessionList := utils.GetSessionList()
	if len(sessionList) > maxCount {
		pterm.Warning.Println(locales.Sprintf("Session count (%d) exceeds the limit (%d)", len(sessionList), maxCount))

		clearAll, err := pterm.DefaultInteractiveConfirm.Show(locales.Sprintf("Clear all sessions?"))
		if err != nil {
			pterm.Error.Println(locales.Sprintf("Confirmation failed:"), err)
			return
		}

		if clearAll {
			utils.ClearAllSessions()
			pterm.Success.Println(locales.Sprintf("All sessions cleared"))
		}
	}
}

func main() {
	utils.EnvInit()
	// 检查是否有init命令
	if len(os.Args) > 1 && os.Args[1] == "init" {
		wizard.RunInitWizard()
		return
	}
	fmt.Println(bergoTitle)
	// 显示完整的版本信息
	versionInfo := version.FormatVersion(version.Version, version.BuildTime, version.CommitHash)
	pterm.Info.Println(fmt.Sprintf("Version %s", versionInfo))
	fmt.Println("\n")

	// 检查更新
	version.CheckAndHandleUpdates()

	readConfig()

	// 检查session数量，如果超过配置的最大值则提示是否清空
	checkAndCleanSessions()

	cli.Debug = config.GlobalConfig.Debug
	mp := agent.NewMainAgent()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mp.Run(ctx, nil)
}
