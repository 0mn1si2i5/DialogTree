// Path: ./cli/ai_cli/chitchat.go

package ai_cli

import (
	"dialogTree/common/cres"
	"dialogTree/service/ai_service"
	"dialogTree/service/redis_service"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"
)

func Chitchat(c *cli.Command) error {
	key := uuid.New().String()

	text := c.String("text")
	if text == "" && c.Args().Len() > 0 {
		text = c.Args().First()
	}

	for {
		cres.Prompt()
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil || input == "exit" {
			redis_service.DelChitChat(key)
			cres.ExitChat()
			break
		}

		err = chat(input, key)
		if err != nil {
			return err
		}
	}
	return nil
}

func chat(input, key string) error {
	field := strconv.Itoa(int(time.Now().Unix()))
	cres.AvatarOnly()
	msg, err := ai_service.PreprocessFromRedis(input, key)
	if err != nil {
		return err
	}
	provider := ai_service.GetDefaultProvider()
	mChan, sChan, err := ai_service.ChatStreamSum(msg, provider)
	if err != nil {
		return err
	}
	record := cres.Stream(mChan)

	var summary string
	for s := range sChan {
		summary += s
	}
	cres.Debug("概要：" + summary)
	redis_service.CacheChitChat(key, field, input, record, summary)
	return nil
}
