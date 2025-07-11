// Path: ./flag/ai_dialog.go

package flag

import "github.com/urfave/cli/v3"

var DialogFlag = []cli.Flag{
	&cli.StringFlag{
		Name:    "text",
		Aliases: []string{"t"}, // 增加 -t 简写
		Usage:   "Text prompt to send",
	},
	&cli.StringFlag{
		Name:    "list",
		Aliases: []string{"l", "li"}, // 增加 -t 简写
		Usage:   "List of all dialogs",
	},
}
