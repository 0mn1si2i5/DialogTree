// Path: ./router/cli_router/web_router.go

package cli_router

import (
	"context"
	"dialogTree/global"
	"dialogTree/router/gin_router"
	"github.com/urfave/cli/v3"
)

var WebUICommand = &cli.Command{
	Name:    "web",
	Aliases: []string{"w"},
	Usage:   "start web ui instead of cli",
	Action: func(ctx context.Context, c *cli.Command) error {
		if global.Config.System.LocalWeb {
			// 前端文件打包进 web
			gin_router.RunWithWeb()
		} else {
			// 开发环境，ide 运行
			gin_router.Run()
		}
		return nil
	},
}
