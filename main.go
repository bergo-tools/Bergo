package main

import (
	"bergo/agent"
	"bergo/config"
	"bergo/locales"
	"bergo/utils"
	"bergo/utils/cli"
	"context"
	"fmt"
	"os"
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

func main() {
	utils.EnvInit()
	readConfig()
	cli.Debug = config.GlobalConfig.Debug
	fmt.Println(bergoTitle)
	mp := agent.NewMainAgent()
	ctx, _ := context.WithCancel(context.Background())
	mp.Run(ctx, nil)
}
