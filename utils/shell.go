package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/aymanbagabas/go-pty"
	"github.com/pterm/pterm"
	"github.com/veops/go-ansiterm"
	"golang.org/x/term"
)

const (
	MAX_OUTPUT_LINE = 5000
)

type Shell struct {
	IsTask bool
}

func (s *Shell) Run(command string) (string, error) {
	if s.IsTask {
		res := s.RunWithTimeout(command)
		if strings.Count(res, "\n") > MAX_OUTPUT_LINE {
			return "", fmt.Errorf("命令执行输出超过%d行,请尝试过滤", MAX_OUTPUT_LINE)
		}
		return res, nil
	}
	result, err := s.RunWithPty(command)
	if err != nil {
		return "", err
	}
	if strings.Count(result, "\n") > MAX_OUTPUT_LINE {
		return "", fmt.Errorf("命令执行输出超过%d行,请尝试过滤", MAX_OUTPUT_LINE)
	}
	return result, nil
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
func (s *Shell) RunWithTimeout(command string) string {
	// 默认超时时间为3min
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	cmd := s.getShellCommand(command)
	cmd.Stdout = nil
	cmd.Stderr = nil

	// 使用带超时的方式执行命令
	var output []byte
	var err error
	done := make(chan struct{})
	go func() {
		output, err = cmd.CombinedOutput()
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			return fmt.Sprintf("cmd error: %v, output: %s", err, string(output))
		}
	case <-ctx.Done():
		// 超时后尝试终止命令
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return "命令执行超过3分钟未返回结果，超时，已终止"
	}

	return strings.TrimSpace(string(output))
}

// 使用伪终端转发过程给用户
func (s *Shell) RunWithPty(command string) (string, error) {
	ptmx, err := pty.New()
	if err != nil {
		return "", err
	}
	defer ptmx.Close()
	height := pterm.GetTerminalHeight() * 3 / 10
	width := pterm.GetTerminalWidth() * 7 / 10
	ptmx.Resize(width, height)

	c := ptmx.Command(`zsh`, "-c", command)
	if er := c.Start(); er != nil {
		return "", er
	}
	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	// NOTE: The goroutine will keep reading until the next keystroke before returning.
	go func() { io.Copy(ptmx, os.Stdin) }()
	buf := bytes.NewBufferString("")
	go func() { io.Copy(io.MultiWriter(os.Stdout, buf), ptmx) }()
	if err := c.Wait(); err != nil {
		return "", err
	}
	return getFinalOutput(buf), nil
}

func getFinalOutput(buff *bytes.Buffer) string {
	width := pterm.GetTerminalWidth() * 7 / 10
	screen := ansiterm.NewScreen(width, MAX_OUTPUT_LINE+2)

	stream := ansiterm.InitByteStream(screen, false)
	defer stream.Close()
	stream.Attach(screen)

	stream.Feed(buff.Bytes())

	strList := screen.Display()
	buf := bytes.NewBufferString("")
	for _, line := range strList {
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	return strings.TrimSpace(buf.String())
}
