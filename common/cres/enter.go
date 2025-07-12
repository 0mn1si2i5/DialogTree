// Path: ./common/cres/enter.go

package cres

import (
	"fmt"
	"time"
)

const (
	agent = "💡Agent"
	user  = "💬 You "
)

func output(object, msg string, newLine bool) {
	var nowStr = time.Now().Format("2006-01-02 15:04:05")
	if newLine {
		msg += "\n"
	}
	out := fmt.Sprintf("[%s]%s: %s", nowStr, object, msg)
	fmt.Print(out)
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

func Stream(msgChan chan string) {
	output(agent, "", false)
	for msg := range msgChan {
		fmt.Print(msg)
	}
	fmt.Println()
}
