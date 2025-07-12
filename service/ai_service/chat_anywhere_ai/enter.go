// Path: ./service/ai_service/chat_anywhere_ai/enter.go

package chat_anywhere_ai

import (
	"dialogTree/global"
	_ "embed"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

const baseURL = "https://api.chatanywhere.tech/v1/chat/completions"

//go:embed prompt_chat.prompt
var chatPrompt string

type AIChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ModelType string

const (
	CA_GPT_35Turbo ModelType = "gpt-3.5-turbo"
	CA_GPT_O1Mini  ModelType = "o1-mini"
	CA_GPT_4O      ModelType = "gpt-4o"
	CA_DS_R1       ModelType = "deepseek-r1"
	CA_DS_V3       ModelType = "deepseek-v3"
	CA_CLA_O4      ModelType = "claude-opus-4-20250514"
	CA_CLA_S4      ModelType = "claude-sonnet-4-20250514"
)

func baseRequest(msg string, model string) (res *http.Response, err error) {
	method := "POST"

	var prompt = chatPrompt

	var m = AIChatRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "system",
				Content: prompt,
			},
			{
				Role:    "user",
				Content: msg,
			},
		},
		Stream: true,
	}
	bd, err := json.Marshal(m)
	if err != nil {
		logrus.Errorf("json 解析失败 %s", err)
		return
	}
	payload := strings.NewReader(string(bd))
	//payload1 := bytes.NewBuffer(bd)

	req, err := http.NewRequest(method, baseURL, payload)
	if err != nil {
		logrus.Errorf("请求解析失败 %s", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+global.Config.Ai.ChatAnywhere.SecretKey)
	req.Header.Add("Content-Type", "application/json")

	res, err = http.DefaultClient.Do(req)
	return
}
