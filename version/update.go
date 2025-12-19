package version

import (
	"bergo/utils/cli"

	"github.com/pterm/pterm"
)

// CheckAndHandleUpdates 检查并处理更新
func CheckAndHandleUpdates() {
	func() {
		hasUpdate, release, err := CheckForUpdates(Version, "bergo-tools", "Bergo")
		if err != nil {
			// 只在调试模式下显示错误
			if cli.Debug {
				pterm.Debug.Println("检查更新失败:", err)
			}
			return
		}
		if !hasUpdate || release == nil {
			return
		}
		pterm.Info.Printf("New version %v available.\n", release.TagName)
		pterm.Info.Printf("%s\n", release.Name)
		pterm.Info.Printf("URL %s\n", release.HTMLURL)
		/*
			if hasUpdate && release != nil {
				handleUpdate(release)
			}
		*/
	}()
}
