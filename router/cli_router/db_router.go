// Path: ./router/cli_router/db_router.go

package cli_router

import (
	"context"
	"dialogTree/core"
	"dialogTree/service/db_service"
	"github.com/urfave/cli/v3"
)

var MigrateDBCommand = &cli.Command{
	Name:    "migrate",
	Aliases: []string{"m", "db"},
	Usage:   "Auto migration of database",
	Action: func(ctx context.Context, c *cli.Command) error {
		core.Init()
		db_service.MigrateDB()
		return nil
	},
}
