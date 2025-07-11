// Path: ./cli/ai_cli/dialog.go

package ai_cli

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v3"
)

var DialogCommand = &cli.Command{
	Name:    "dialog",
	Aliases: []string{"d"},
	Usage:   "Interact with persistent dialog sessions",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "text",
			Aliases: []string{"t"}, // 增加 -t 简写
			Usage:   "Text prompt to send",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		text := c.String("text")
		if text == "" && c.Args().Len() > 0 {
			text = c.Args().First()
		}
		// 模拟 GPT 响应
		fmt.Printf("🤖 GPT: [Dialog] You said: %s\n", text)
		return nil
	},
}
