// Path: ./cli/ai_cli/dialog.go

package ai_cli

import (
	"context"
	"dialogTree/core"
	"dialogTree/service/tea_service"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v3"
)

func ShowDialogs(ctx context.Context, c *cli.Command) error {
	// 1. 假设 dialogs 是你查出来的对话列表
	core.CoreInit()
	p, err := tea_service.ShowAllSessions()
	if err != nil {
		return err
	}
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func Enter(dialog string) error {
	fmt.Printf("现在进入对话：%s\n", dialog)
	// 这里可以实现聊天循环、显示树、再选择等
	// 例如进入聊天模式
	for {
		fmt.Print("你：")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil || input == "exit" {
			fmt.Println("退出对话。")
			break
		}
		// 这里可以调用 AI 回复
		fmt.Printf("AI：你说了 %s\n", input)
	}
	return nil
}

func EnterDialog(ctx context.Context, c *cli.Command) error {
	// 这里只处理 enter 逻辑
	fmt.Print("输入编号：")
	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len("TODO!!!!!!!!!!!!") {
		fmt.Println("输入无效")
		return nil
	}
	return Enter(strconv.Itoa(choice))
}

func EnterRecent(ctx context.Context, c *cli.Command) error {
	// 这里只处理 chat 逻辑
	fmt.Println("最近的一次对话")
	return nil
}
