// Path: ./service/ai_service/common/client.go

package common

import (
	"dialogTree/service/ai_service/prompts"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// UniversalChatRequest 通用的聊天请求结构
type UniversalChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIProviderConfig AI提供商配置
type AIProviderConfig struct {
	BaseURL   string
	APIKey    string
	Model     string
}

// MakeRequest 通用的HTTP请求函数
func MakeRequest(config AIProviderConfig, msg string, summarize bool) (res *http.Response, err error) {
	method := "POST"

	// 选择prompt
	var prompt = prompts.ChatPrompt
	if summarize {
		prompt = prompts.SummarizePrompt
	}

	// 构建请求体
	requestBody := UniversalChatRequest{
		Model: config.Model,
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

	// 序列化请求体
	bd, err := json.Marshal(requestBody)
	if err != nil {
		logrus.Errorf("json 解析失败 %s", err)
		return
	}
	payload := strings.NewReader(string(bd))

	// 创建HTTP请求
	req, err := http.NewRequest(method, config.BaseURL, payload)
	if err != nil {
		logrus.Errorf("请求解析失败 %s", err)
		return
	}

	// 设置请求头
	req.Header.Add("Authorization", "Bearer "+config.APIKey)
	req.Header.Add("Content-Type", "application/json")

	// 发送请求
	res, err = http.DefaultClient.Do(req)
	return
}