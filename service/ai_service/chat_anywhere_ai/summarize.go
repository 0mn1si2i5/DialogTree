// Path: ./service/ai_service/chat_anywhere_ai/summarize.go

package chat_anywhere_ai

import (
	"dialogTree/global"
	"dialogTree/models"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

// Summarize0 没用了，更耗费 token
func Summarize0(msg string) (resp string, err error) {
	res, err := baseRequest(msg, global.Config.Ai.BackendAi.Model, true)
	if err != nil {
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("响应读取失败 %s", err)
		return
	}

	var aiRes AIChatResponse
	err = json.Unmarshal(body, &aiRes)
	if err != nil {
		logrus.Errorf("响应解析失败 %s\n原始数据 %s", err, string(body))
	}

	return aiRes.Choices[0].Message.Content, nil
}

func ChatWithSummarize(msg string, parentID uint) (msgChan, sumChan chan string, err error) {
	sendMsg := fmt.Sprintf("¥Q:%s;", msg)
	if parentID != 0 {
		var msgModel models.MessageModel
		err = global.DB.Find(&msgModel, "id = ?", parentID).Preload("ParentModel").Preload("ParentModel.ParentModel").Error
		if err != nil {
			fmt.Println(err)
			return
		}
		var q3, a3, q2, a2, q1, a1 string
		q1, a1 = msgModel.Prompt, msgModel.Answer
		if msgModel.ParentModel != nil {
			q2 = msgModel.ParentModel.Prompt
			a2 = msgModel.ParentModel.Answer

			if msgModel.ParentModel.ParentModel != nil {
				q3 = msgModel.ParentModel.ParentModel.Prompt
				a3 = msgModel.ParentModel.ParentModel.Answer
			}
		}
		sendMsg = fmt.Sprintf("¥H:%s;¥3Q:%s;¥3A:%s;¥2Q:%s;¥2A:%s;¥1Q:%s;¥1A:%s;¥Q:%s;",
			msgModel.Summary, q3, a3, q2, a2, q1, a1, msg,
		)
	}
	return ChatStreamSum(sendMsg)
}
