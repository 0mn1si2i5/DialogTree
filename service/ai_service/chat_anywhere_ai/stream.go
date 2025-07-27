// Path: ./service/ai_service/chat_anywhere_ai/stream.go

package chat_anywhere_ai

import (
	"bufio"
	"dialogTree/common/cres"
	"dialogTree/global"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
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

func stream(scanner *bufio.Scanner, res *http.Response, msgChan chan string) {
	defer close(msgChan)
	defer res.Body.Close()

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
			return
		}

		// 解析 json 数据
		var aiRes AIChatStreamResponse
		err := json.Unmarshal([]byte(jsonData), &aiRes)
		if err != nil {
			logrus.Errorf("JSON 解析失败: %v\n原始数据: %s", err, jsonData)
			continue
		}

		if len(aiRes.Choices) == 0 {
			continue
		}

		content := aiRes.Choices[0].Delta.Content
		if content == "" {
			continue
		}

		msgChan <- content
	}
}

func streamSplitter(scanner *bufio.Scanner, res *http.Response, msgChan, sumChan chan string) {
	defer close(sumChan)
	defer res.Body.Close()

	var slidingBuffer strings.Builder // 缓冲所有 token
	const marker = "^¥&"

	var state int8 = 0 // 状态机：0=正常，1=^，2=^¥，3=^¥&

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
			wholeMsg := slidingBuffer.String()
			msgs := strings.SplitN(wholeMsg, marker, 2)

			if len(msgs) == 2 {
				summary := msgs[1]
				sumChan <- summary // ✅ 会阻塞等待消费
			} else {
				logrus.Warn("\n未能正确提取摘要")
				close(msgChan)
				return
			}
			cres.Debug("完整消息：" + wholeMsg)

			_, ok := <-msgChan
			if ok {
				close(msgChan)
			}
			return
		}

		// 解析 json 数据
		var aiRes AIChatStreamResponse
		err := json.Unmarshal([]byte(jsonData), &aiRes)
		if err != nil {
			logrus.Errorf("JSON 解析失败: %v\n原始数据: %s", err, jsonData)
			continue
		}

		if len(aiRes.Choices) == 0 {
			continue
		}

		content := aiRes.Choices[0].Delta.Content
		if content == "" {
			continue
		}

		// 组装缓冲内容
		slidingBuffer.WriteString(content)

		switch state {
		case 0:
			if content == "^" || strings.HasSuffix(slidingBuffer.String(), "^") {
				state = 1
			} else if content == "^¥" || strings.HasSuffix(slidingBuffer.String(), "^¥") {
				state = 2
			} else if content == marker || strings.Contains(slidingBuffer.String(), marker) {
				state = 3
			} else {
				msgChan <- content
			}
		case 1:
			if content == "¥" || strings.HasSuffix(slidingBuffer.String(), "^¥") {
				state = 2
			} else {
				msgChan <- content
			}
		case 2:
			if content == "&" || strings.Contains(slidingBuffer.String(), marker) {
				state = 3
				close(msgChan)
			} else {
				msgChan <- content
			}
		}
	}
}

func ChatStreamSum(msg string) (msgChan, sumChan chan string, err error) {
	// 检查AI配置密钥，如果为空则返回模拟响应
	if global.Config.Ai.ChatAnywhere.SecretKey == "" {
		logrus.Info("AI密钥为空，返回模拟响应用于测试")
		
		msgChan = make(chan string)
		sumChan = make(chan string)
		
		// 启动goroutine发送模拟响应
		go func() {
			// 模拟AI回答
			mockAnswer := "这是一个模拟的AI回答，用于测试分叉功能。"
			for _, char := range mockAnswer {
				msgChan <- string(char)
			}
			
			// 关闭msgChan，模拟消息结束
			close(msgChan)
			
			// 模拟摘要JSON
			mockSummary := `{"title": "测试对话", "summary": "这是一个测试摘要"}`
			sumChan <- mockSummary
			close(sumChan)
		}()
		
		return
	}

	res, err := baseRequest(msg, global.Config.Ai.ChatAnywhere.Model, true)
	if err != nil {
		return
	}

	msgChan = make(chan string)
	sumChan = make(chan string)

	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)

	go streamSplitter(scanner, res, msgChan, sumChan)

	return
}

func ChatStream(msg string) (msgChan chan string, err error) {
	res, err := baseRequest(msg, global.Config.Ai.ChatAnywhere.Model, false)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		if res.StatusCode == 429 {
			err = errors.New("请求过于频繁，请稍后重试")
		} else {
			err = errors.New(fmt.Sprintf("服务器响应错误 %d", res.StatusCode))
		}
		return
	}

	msgChan = make(chan string)

	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)

	go stream(scanner, res, msgChan)

	return
}
