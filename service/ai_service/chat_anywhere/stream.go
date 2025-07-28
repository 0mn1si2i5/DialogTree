// Path: ./service/ai_service/chat_anywhere/stream.go

package chat_anywhere

import (
	"dialogTree/global"
	"dialogTree/service/ai_service/common"
	"github.com/sirupsen/logrus"
)

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

	config := getConfig()
	return common.CreateChatStreamWithSummary(config, msg)
}

func ChatStream(msg string) (msgChan chan string, err error) {
	config := getConfig()
	return common.CreateChatStream(config, msg)
}
