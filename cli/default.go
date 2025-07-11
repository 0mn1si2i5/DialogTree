// Path: ./cli/default.go

package cliapp

import (
	"bufio"
	"context"
	"fmt"
	"github.com/urfave/cli/v3"
	"io"
	"os"
	"strings"
)

func Default() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		var input string

		// 1️⃣ 如果来自管道（stdin is not terminal）
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// 从管道读取
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			input = string(data)
		} else {
			// 2️⃣ 否则，从命令行参数读取
			args := cmd.Args().Slice()
			if len(args) == 0 {
				// 3️⃣ 没有参数，也没有管道输入：给用户提示
				fmt.Println("🤖 Enter prompt:")
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
			fmt.Println("error: No prompt provided.")
			return nil
		}

		// 模拟 GPT 响应
		fmt.Printf("🤖 GPT: [Chitchat] You said: %s\n", input)
		return nil
	}
}
