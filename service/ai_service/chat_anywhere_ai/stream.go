// Path: ./service/ai_service/chat_anywhere_ai/stream.go

package chat_anywhere_ai

import (
	"bufio"
	"dialogTree/global"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"strings"
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

type AIChatStreamResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		Logprobs     any    `json:"logprobs"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
}

func ChatStream(msg string) (msgChan chan string, err error) {
	res, err := baseRequest(msg, global.Config.Ai.ChatAnywhere.Model)
	if err != nil {
		return
	}

	msgChan = make(chan string)

	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			// 跳过空行
			if line == "" {
				continue
			}

			// 检查是否是 SSE 数据行
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			// 提取 JSON 部分（去掉 "data: " 前缀）
			jsonData := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标记
			if jsonData == "[DONE]" {
				close(msgChan)
				return
			}

			// 解析 json 数据
			var aiRes AIChatStreamResponse
			err = json.Unmarshal([]byte(jsonData), &aiRes)
			if err != nil {
				logrus.Errorf("JSON 解析失败: %v\n原始数据: %s", err, jsonData)
				continue
			}

			if len(aiRes.Choices) == 0 {
				continue
			}

			msgChan <- aiRes.Choices[0].Delta.Content
		}
	}()

	return
}
