package main

import (
	"bergo/utils"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/aymanbagabas/go-pty"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

type PTY interface {
	Resize(w, h int) error
}

func test() (string, error) {
	ptmx, err := pty.New()
	if err != nil {
		return "", err
	}
	defer ptmx.Close()
	height := pterm.GetTerminalHeight() * 3 / 10
	width := pterm.GetTerminalWidth() * 7 / 10
	ptmx.Resize(width, height)

	c := ptmx.Command(`zsh`, "-c", "ls -al")
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
	return buf.String(), nil
}

func main() {
	s := utils.Shell{
		IsTask: false,
	}
	output, err := s.Run("ls -l")
	if err != nil {
		fmt.Println("Run command failed:", err)
	}
	fmt.Println(output)

	s2 := utils.Shell{
		IsTask: false,
	}
	output2, err := s2.Run("grep main main.go")
	if err != nil {
		fmt.Println("Run command failed:", err)
	}
	fmt.Println(output2)
}
