// Path: ./core/enter.go

package core

import (
	"dialogTree/global"
	"dialogTree/router/cli_router"
	"os"
)

func QuickChat() {
	if len(os.Args) == 1 || (len(os.Args) > 1 && (os.Args[1] == "chitchat" || os.Args[1] == "c")) {
		global.Config = ReadConf(true)
		cli_router.Run()
		os.Exit(1)
	}
}
