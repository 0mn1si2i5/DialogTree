// Path: ./router/cli_router/web_router.go

package cli_router

import (
	"context"
	"dialogTree/router/gin_router"
	"github.com/urfave/cli/v3"
)

var WebUICommand = &cli.Command{
	Name:    "web",
	Aliases: []string{"w"},
	Usage:   "start web ui instead of cli",
	Action: func(ctx context.Context, c *cli.Command) error {
		gin_router.Run()
		return nil
	},
}
