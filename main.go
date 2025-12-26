package main

import (
	"bergo/agent"
	"bergo/config"
	"bergo/locales"
	"bergo/skills"
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

// checkAndCleanSessions 检查session数量，如果超过配置的最大值则自动清理旧session
func checkAndCleanSessions() {
	maxCount := config.GlobalConfig.MaxSessionCount
	// 如果配置值为0，使用默认值30
	if maxCount <= 0 {
		maxCount = 30
	}

	sessionList := utils.GetSessionList()
	if len(sessionList) > maxCount {
		// 保留最新的 maxCount 个 session（列表末尾是最新的）
		keepList := sessionList[len(sessionList)-maxCount:]
		removedCount := len(sessionList) - maxCount
		utils.SetSessionList(keepList)
		pterm.Info.Println(locales.Sprintf("Cleaned %d old sessions, keeping %d", removedCount, maxCount))
	}
}

// checkAndRecoverSession 检查是否存在 memento 文件，如果存在说明上次异常退出
// 返回需要恢复的 sessionId，如果不需要恢复则返回空字符串
func checkAndRecoverSession() string {
	mementoPath := "./.bergo.memento"
	if _, err := os.Stat(mementoPath); os.IsNotExist(err) {
		return ""
	}

	// 存在 memento 文件，说明上次异常退出
	pterm.Warning.Println(locales.Sprintf("Detected abnormal exit from last session"))

	// 获取最后一个 session
	sessionList := utils.GetSessionList()
	if len(sessionList) == 0 {
		// 没有 session 记录，删除 memento 文件
		os.Remove(mementoPath)
		return ""
	}

	lastSession := sessionList[len(sessionList)-1]
	pterm.Info.Println(locales.Sprintf("Last session: %s", lastSession.SessionId))

	// 询问用户是否恢复
	recover, err := pterm.DefaultInteractiveConfirm.Show(locales.Sprintf("Recover last session and revert to last checkpoint?"))
	if err != nil {
		pterm.Error.Println(locales.Sprintf("Confirmation failed:"), err)
		os.Remove(mementoPath)
		return ""
	}

	if !recover {
		// 用户选择不恢复，删除 memento 文件
		os.Remove(mementoPath)
		pterm.Info.Println(locales.Sprintf("Session recovery cancelled"))
		return ""
	}

	// 用户选择恢复，执行回退操作
	timeline := &utils.Timeline{}
	timeline.Init(lastSession.SessionId)
	timeline.Load()
	timeline.RevertToLastCheckpoint()

	pterm.Success.Println(locales.Sprintf("Session recovered and reverted to last checkpoint"))
	return lastSession.SessionId
}
func loadSkills() {
	manager := skills.GetManager()
	// 先释放内置skills到用户主目录
	_ = manager.LoadBuiltinSkills()
	// 从用户主目录加载所有skills（包括内置和用户自定义的）
	_ = manager.LoadSkills()
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

	pterm.Info.Println(fmt.Sprintf("Version %s", version.FormatVersion(version.Version, version.BuildTime, version.CommitHash)))
	fmt.Println("\n")
	// 加载 skills
	loadSkills()

	// 检查更新
	version.CheckAndHandleUpdates()

	readConfig()

	// 检查session数量，如果超过配置的最大值则提示是否清空
	checkAndCleanSessions()

	// 检查是否存在 memento 文件，如果存在说明上次异常退出
	recoverySessionId := checkAndRecoverSession()

	cli.Debug = config.GlobalConfig.Debug
	mp := agent.NewMainAgent()

	// 如果有需要恢复的 session，设置恢复标记
	if recoverySessionId != "" {
		mp.SetRecoverySession(recoverySessionId)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mp.Run(ctx, nil)
}
