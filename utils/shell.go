package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Shell struct {
}

// getShellCommand 根据操作系统返回合适的shell命令
func (s *Shell) getShellCommand(command string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		// Windows 系统使用 PowerShell
		return exec.Command("powershell", "-Command", command)
	case "darwin":
		// macOS 系统，优先使用 zsh，如果不存在则使用 bash
		if _, err := exec.LookPath("zsh"); err == nil {
			return exec.Command("zsh", "-c", command)
		}
		return exec.Command("bash", "-c", command)
	case "linux":
		// Linux 系统，优先使用 bash，如果不存在则使用 sh
		if _, err := exec.LookPath("bash"); err == nil {
			return exec.Command("bash", "-c", command)
		}
		return exec.Command("sh", "-c", command)
	default:
		// 其他系统默认使用 sh
		return exec.Command("sh", "-c", command)
	}
}

// Run 执行shell命令，根据操作系统自动选择合适的shell
func (s *Shell) Run(command string) string {
	cmd := s.getShellCommand(command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("cmd error: %v, output: %s", err, string(output))
	}
	return strings.TrimSpace(string(output))
}

// RunWithShell 指定特定的shell执行命令
func (s *Shell) RunWithShell(shell, command string) string {
	var cmd *exec.Cmd

	switch strings.ToLower(shell) {
	case "bash":
		cmd = exec.Command("bash", "-c", command)
	case "zsh":
		cmd = exec.Command("zsh", "-c", command)
	case "sh":
		cmd = exec.Command("sh", "-c", command)
	case "powershell", "pwsh":
		cmd = exec.Command("powershell", "-Command", command)
	case "cmd":
		cmd = exec.Command("cmd", "/c", command)
	default:
		// 如果指定的shell不存在，回退到自动选择
		return s.Run(command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("cmd error: %v, output: %s", err, string(output))
	}
	return strings.TrimSpace(string(output))
}
