// Path: ./service/ai_service/enter.go

package ai_service

import (
	"dialogTree/common/cres"
	"dialogTree/global"
	"dialogTree/service/ai_service/chat_anywhere"
	"dialogTree/service/ai_service/deepseek"
	"dialogTree/service/ai_service/openai"
	"dialogTree/service/redis_service"
	"fmt"
	"sort"
)

type RequestType int8

const (
	ChatAiRequest      RequestType = 1
	SummarizeAiRequest RequestType = 2
)

type AIProvider string

const (
	ChatAnywhereProvider AIProvider = "chatanywhere"
	OpenAIProvider       AIProvider = "openai"
	DeepSeekProvider     AIProvider = "deepseek"
	BackendAIProvider    AIProvider = "backendai"
)

// ChatStreamSum 统一的流式聊天+摘要接口
func ChatStreamSum(msg string, provider AIProvider) (msgChan, sumChan chan string, err error) {
	switch provider {
	case ChatAnywhereProvider:
		return chat_anywhere.ChatStreamSum(msg)
	case DeepSeekProvider:
		return deepseek.ChatStreamSum(msg)
	case OpenAIProvider:
		return openai.ChatStreamSum(msg)
	case BackendAIProvider:
		// BackendAI使用ChatAnywhere的实现，但使用不同的配置
		return chat_anywhere.ChatStreamSum(msg)
	default:
		// 默认使用ChatAnywhere
		return chat_anywhere.ChatStreamSum(msg)
	}
}

// ChatStream 统一的流式聊天接口
func ChatStream(msg string, provider AIProvider) (msgChan chan string, err error) {
	switch provider {
	case ChatAnywhereProvider:
		return chat_anywhere.ChatStream(msg)
	case DeepSeekProvider:
		return deepseek.ChatStream(msg)
	case OpenAIProvider:
		return openai.ChatStream(msg)
	case BackendAIProvider:
		return chat_anywhere.ChatStream(msg)
	default:
		return chat_anywhere.ChatStream(msg)
	}
}

// GetDefaultProvider 根据配置获取默认的AI提供商
func GetDefaultProvider() AIProvider {
	// 优先级：ChatAnywhere > DeepSeek > BackendAI > OpenAI
	if global.Config.Ai.ChatAnywhere.SecretKey != "" {
		return ChatAnywhereProvider
	}
	if global.Config.Ai.DeepSeek.SecretKey != "" {
		return DeepSeekProvider
	}
	if global.Config.Ai.BackendAi.SecretKey != "" {
		return BackendAIProvider
	}
	if global.Config.Ai.OpenAI.SecretKey != "" {
		return OpenAIProvider
	}
	return ChatAnywhereProvider
}

// PreprocessFromRedis 从Redis预处理消息（通用函数）
func PreprocessFromRedis(msg, key string) (processedMsg string, err error) {
	pmap, amap, summary := redis_service.GetChitChat(key)
	processedMsg += fmt.Sprintf("¥H:%s;", summary)
	prompts, answers := sortMap(pmap), sortMap(amap)
	for i := range prompts {
		processedMsg += fmt.Sprintf("¥%dQ:%s;¥%dA:%s;", len(prompts)-i, prompts[i], len(prompts)-i, answers[i])
	}
	processedMsg += fmt.Sprintf("¥Q:%s;", msg)
	cres.Debug("\n" + processedMsg + "\n")
	return
}

func sortMap(m map[string]string) (values []string) {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		values = append(values, m[k])
	}
	return
}
