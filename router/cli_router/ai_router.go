// Path: ./router/cli_router/ai_router.go

package cli_router

import (
	"context"
	"dialogTree/cli/ai_cli"
	"dialogTree/common/cres"
	"dialogTree/core"
	"dialogTree/flag"
	"dialogTree/global"
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
	Action: func(ctx context.Context, c *cli.Command) (err error) {
		cres.Debug("=== 进入 chitchat 模式 ===")
		global.Redis = core.InitRedis(true)
		err = ai_cli.Chitchat(c)
		return
	},
}

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
			Action:  ai_cli.ShowDialogs,
		},
		{
			Name:    "enter",
			Aliases: []string{"e", "en", "i", "in"},
			Usage:   "Enter a certain dialog",
			Flags:   flag.DialogFlag, // 这里也可以用同一组 flag
			Action:  ai_cli.EnterDialog,
		},
		{
			Name:    "recent",
			Aliases: []string{"r", "re", "c", "ch"},
			Usage:   "Enter the most recent dialog",
			Flags:   flag.DialogFlag,
			Action:  ai_cli.EnterRecent,
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		cres.Debug("=== 进入 dialog 模式 ===")
		core.CoreInit()
		if c.Args().Len() == 0 && len(c.FlagNames()) == 0 {
			return ai_cli.EnterRecent(ctx, c)
		}
		return cli.ShowSubcommandHelp(c)
	},
}
