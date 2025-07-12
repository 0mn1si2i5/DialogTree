// Path: ./common/cres/enter.go

package cres

import (
	"dialogTree/global"
	"fmt"
	"time"
)

const (
	agent = "💡"
	user  = "💬"
)

func Output(msg string) {
	var model = global.Config.Ai.ChatAnywhere.Model
	var nowStr = time.Now().Format("2006-01-02 15:04:05")
	output := fmt.Sprintf("[%s|%s] %s %s\n", nowStr, model, agent, msg)
	fmt.Print(output)
}

func Prompt() {
	var nowStr = time.Now().Format("2006-01-02 15:04:05")
	output := fmt.Sprintf("[%s|%s] %s: ", nowStr, "user", user)
	fmt.Print(output)
}
