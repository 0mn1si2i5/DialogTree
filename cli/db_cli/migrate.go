// Path: ./cli/db_cli/migrate.go

package db_cli

import (
	"context"
	"dialogTree/service/db_service"
	"github.com/urfave/cli/v3"
)

var MigrateDBCommand = &cli.Command{
	Name:    "migrate",
	Aliases: []string{"m", "db"},
	Usage:   "Auto migration of database",
	Action: func(ctx context.Context, c *cli.Command) error {
		db_service.MigrateDB()
		return nil
	},
}
