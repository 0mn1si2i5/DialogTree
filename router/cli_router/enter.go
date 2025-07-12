// Path: ./router/cli_router/enter.go

package cli_router

import (
	"context"
	"dialogTree/cli/ai_cli"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"os"
)

func Run() {
	app := App
	if err := app.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}

var App = &cli.Command{
	Name:  "dialogtree",
	Usage: "Manage structured dialogs from CLI",
	Commands: []*cli.Command{
		ChitchatCommand,
		DialogCommand,
		MigrateDBCommand,
		WebUICommand,
	},
	Action: ai_cli.Default,
}
