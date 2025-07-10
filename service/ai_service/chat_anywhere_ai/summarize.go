// Path: ./service/ai_service/chat_anywhere_ai/summarize.go

package chat_anywhere_ai

import (
	"dialogTree/global"
	"dialogTree/service/ai_service"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
)

type AIChatResponse struct {
	Id      string `json:"id"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Object  string `json:"object"`
	Usage   struct {
		PromptTokens            int `json:"prompt_tokens"`
		CompletionTokens        int `json:"completion_tokens"`
		TotalTokens             int `json:"total_tokens"`
		CompletionTokensDetails struct {
			AudioTokens     int `json:"audio_tokens"`
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
		PromptTokensDetails struct {
			AudioTokens  int `json:"audio_tokens"`
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage"`
	SystemFingerprint interface{} `json:"system_fingerprint"`
}

func Summarize(msg string) (resp string, err error) {
	res, err := baseRequest(msg, global.Config.Ai.BackendAi.Model, ai_service.SummarizeAiRequest)
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
