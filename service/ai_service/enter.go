// Path: ./service/ai_service/enter.go

package ai_service

type RequestType int8

const (
	ChatAiRequest      RequestType = 1
	SummarizeAiRequest RequestType = 2
)
