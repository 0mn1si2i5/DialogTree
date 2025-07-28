// Path: ./service/ai_service/openai/summarize.go

package openai

import (
	"dialogTree/service/ai_service/common"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
)

// Summarize0 没用了，更耗费 token
func Summarize0(msg string) (resp string, err error) {
	config := getConfig()
	res, err := common.MakeRequest(config, msg, true)
	if err != nil {
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("响应读取失败 %s", err)
		return
	}

	var aiRes common.UniversalChatResponse
	err = json.Unmarshal(body, &aiRes)
	if err != nil {
		logrus.Errorf("响应解析失败 %s\n原始数据 %s", err, string(body))
	}

	return aiRes.Choices[0].Message.Content, nil
}