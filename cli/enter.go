// Path: ./cli/enter.go

package cliapp

import (
	"dialogTree/cli/ai_cli"
	"dialogTree/cli/db_cli"

	"github.com/urfave/cli/v3"
)

var App = &cli.Command{
	Name:  "dialogtree",
	Usage: "Manage structured dialogs from CLI",
	Commands: []*cli.Command{
		ai_cli.ChitchatCommand,
		ai_cli.DialogCommand,
		db_cli.MigrateDBCommand,
	},
	Action: Default(),
}
