// Path: ./cli/ai_cli/chitchat.go

package ai_cli

import (
	"context"
	"dialogTree/service/ai_service/chat_anywhere_ai"
	"fmt"

	"github.com/urfave/cli/v3"
)

var ChitchatCommand = &cli.Command{
	Name:    "chitchat",
	Aliases: []string{"c", "chat"},
	Usage:   "Quick one-off chat (non-blocking)",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "text",
			Aliases: []string{"t"}, // 增加 -t 简写
			Usage:   "Text prompt to send",
		},
	},
	Action: Chitchat,
}

func Chitchat(ctx context.Context, c *cli.Command) error {
	text := c.String("text")
	if text == "" && c.Args().Len() > 0 {
		text = c.Args().First()
	}

	for {
		fmt.Print("你：")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil || input == "exit" {
			fmt.Println("退出对话。")
			break
		}
		// 这里可以调用 AI 回复
		err = chat(input)
		if err != nil {
			return err
		}
	}
	return nil
}

func chat(input string) error {
	//text := c.String("text")
	//if text == "" && c.Args().Len() > 0 {
	//	text = c.Args().First()
	//}
	// 响应
	fmt.Printf("🤖 Agent: ")
	mChan, sChan, err := chat_anywhere_ai.ChatWithSummarize(input, 0) // todo parent
	if err != nil {
		return err
	}
	for m := range mChan {
		fmt.Print(m)
	}

	var summary string
	for s := range sChan {
		fmt.Println(s)
		summary += s
	}
	fmt.Println("summary done: ", summary)

	return nil
}
