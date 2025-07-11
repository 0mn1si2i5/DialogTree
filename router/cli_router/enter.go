// Path: ./router/cli_router/enter.go

package cli_router

import (
	"context"
	cliapp "dialogTree/cli"
	"github.com/sirupsen/logrus"
	"os"
)

func Run() {
	app := cliapp.App
	if err := app.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}
