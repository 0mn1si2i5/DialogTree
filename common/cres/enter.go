// Path: ./common/cres/enter.go

package cres

import (
	"dialogTree/global"
	"fmt"
	"time"
)

var agent string
var user = "💬 You "

func SetAgentLabel() {
	if global.Config == nil {
		agent = "💡Agent"
	} else {
		agent = "💡" + global.Config.Ai.ChatAnywhere.Model
	}
}

func output(object, msg string, newLine bool) {
	var nowStr = time.Now().Format("2006-01-02 15:04:05")
	if newLine {
		msg += "\n"
	}
	out := fmt.Sprintf("[%s]%s: %s", nowStr, object, msg)
	fmt.Print(out)
}

func AvatarOnly() {
	output(agent, "", false)
}

func Output(msg string) {
	output(agent, msg, false)
}

func Prompt() {
	output(user, "", false)
}

func Error(err error) {
	ErrorMsg(err.Error())
}

func ErrorMsg(msg string) {
	output(" error", msg, true)
}

func Stream(msgChan chan string) (record string) {
	for s := range msgChan {
		fmt.Print(s)
		record += s
	}
	fmt.Println()
	return
}

func ExitChat() {
	output("exit", "本次会话结束，再见！", true)
}

func Debug(msg string) {
	if global.Config.System.Mode == "debug" {
		output("debug", msg, true)
	}
}
