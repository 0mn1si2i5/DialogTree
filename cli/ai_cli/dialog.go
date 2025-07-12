// Path: ./cli/ai_cli/dialog.go

package ai_cli

import (
	"context"
	"dialogTree/core"
	"dialogTree/flag"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v3"
)

var DialogCommand = &cli.Command{
	Name:    "dialog",
	Aliases: []string{"d"},
	Usage:   "Interact with persistent dialog sessions",
	Commands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l", "ls", "li", "show"},
			Usage:   "Show all the dialogs",
			Flags:   flag.DialogFlag, // 这里可以只用需要的 flag
			Action:  showDialogs,
		},
		{
			Name:    "enter",
			Aliases: []string{"e", "en", "i", "in"},
			Usage:   "Enter a certain dialog",
			Flags:   flag.DialogFlag, // 这里也可以用同一组 flag
			Action:  enterDialog,
		},
		{
			Name:    "recent",
			Aliases: []string{"r", "re", "c", "ch"},
			Usage:   "Enter the most recent dialog",
			Flags:   flag.DialogFlag,
			Action:  enterRecent,
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		fmt.Println("=== 进入 dialog 模式 ===")
		core.CoreInit()
		if c.Args().Len() == 0 && len(c.FlagNames()) == 0 {
			return enterRecent(ctx, c)
		}
		return cli.ShowSubcommandHelp(c)
	},
}

// 各自的处理函数
func showDialogs(ctx context.Context, c *cli.Command) error {
	// 1. 假设 dialogs 是你查出来的对话列表
	dialogs := []string{"对话1", "对话2", "对话3"}
	fmt.Println("请选择一个对话：")
	for i, d := range dialogs {
		fmt.Printf("%d. %s\n", i+1, d)
	}

	// 2. 阻塞等待用户输入
	fmt.Print("输入编号：")
	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(dialogs) {
		fmt.Println("输入无效")
		return nil
	}

	// 3. 根据选择处理
	selected := dialogs[choice-1]
	fmt.Printf("你选择了：%s\n", selected)
	// 这里可以进入后续对话逻辑，比如进入 chat 模式
	return enter(selected)
}

func enter(dialog string) error {
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

func enterDialog(ctx context.Context, c *cli.Command) error {
	// 这里只处理 enter 逻辑
	fmt.Print("输入编号：")
	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len("TODO!!!!!!!!!!!!") {
		fmt.Println("输入无效")
		return nil
	}
	return enter(strconv.Itoa(choice))
}

func enterRecent(ctx context.Context, c *cli.Command) error {
	// 这里只处理 chat 逻辑
	fmt.Println("最近的一次对话")
	return nil
}
