// Path: ./cli/ai_cli/chitchat.go

package ai_cli

import (
	"context"
	"dialogTree/common/cres"
	"dialogTree/service/ai_service/chat_anywhere_ai"
	"fmt"

	"github.com/urfave/cli/v3"
)

func Chitchat(ctx context.Context, c *cli.Command) error {
	text := c.String("text")
	if text == "" && c.Args().Len() > 0 {
		text = c.Args().First()
	}

	for {
		cres.Prompt()
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil || input == "exit" {
			cres.ExitChat()
			break
		}

		err = chat(input)
		if err != nil {
			return err
		}
	}
	return nil
}

func chat(input string) error {
	cres.AvatarOnly()
	msg, err := chat_anywhere_ai.PreprocessContext(input, 0) // todo parent
	if err != nil {
		return err
	}
	mChan, sChan, err := chat_anywhere_ai.ChatStreamSum(msg)
	if err != nil {
		return err
	}
	cres.Stream(mChan)

	var summary string
	for s := range sChan {
		fmt.Println(s)
		summary += s
	}
	fmt.Println("summary done: ", summary)

	return nil
}
