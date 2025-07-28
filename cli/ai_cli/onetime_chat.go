// Path: ./cli/ai_cli/onetime_chat.go

package ai_cli

import (
	"bufio"
	"context"
	"dialogTree/common/cres"
	"dialogTree/service/ai_service"
	"github.com/urfave/cli/v3"
	"io"
	"os"
	"strings"
)

func OneTimeChat(ctx context.Context, cmd *cli.Command) error {
	var input string

	// 1-如果来自管道（stdin is not terminal）
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// 从管道读取
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		input = string(data)
	} else {
		// 2-否则，从命令行参数读取
		args := cmd.Args().Slice()
		if len(args) == 0 {
			// 3-没有参数，也没有管道输入：给用户提示
			cres.Prompt()
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				input = scanner.Text()
			}
		} else {
			input = strings.Join(args, " ")
		}
	}

	input = strings.TrimSpace(input)
	if input == "" {
		cres.ErrorMsg("No prompt provided")
		return nil
	}

	cres.AvatarOnly()
	provider := ai_service.GetDefaultProvider()
	msgChan, err := ai_service.ChatStream(input, provider)
	if err != nil {
		return err
	}
	cres.Stream(msgChan)
	return nil
}
