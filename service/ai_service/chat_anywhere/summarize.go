// Path: ./service/ai_service/chat_anywhere/summarize.go

package chat_anywhere

import (
	"dialogTree/common/cres"
	"dialogTree/service/ai_service/common"
	"dialogTree/service/redis_service"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/sirupsen/logrus"
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

//func PreprocessFromSQL(msg string, parentID uint) (processedMsg string, err error) {
//	processedMsg = fmt.Sprintf("¥Q:%s;", msg)
//	if parentID != 0 {
//		var msgModel models.MessageModel
//		err = global.DB.Find(&msgModel, "id = ?", parentID).Preload("ParentModel").Preload("ParentModel.ParentModel").Error
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		var q3, a3, q2, a2, q1, a1 string
//		q1, a1 = msgModel.Prompt, msgModel.Answer
//		if msgModel.ParentModel != nil {
//			q2 = msgModel.ParentModel.Prompt
//			a2 = msgModel.ParentModel.Answer
//
//			if msgModel.ParentModel.ParentModel != nil {
//				q3 = msgModel.ParentModel.ParentModel.Prompt
//				a3 = msgModel.ParentModel.ParentModel.Answer
//			}
//		}
//		processedMsg = fmt.Sprintf("¥H:%s;¥3Q:%s;¥3A:%s;¥2Q:%s;¥2A:%s;¥1Q:%s;¥1A:%s;¥Q:%s;",
//			msgModel.Summary, q3, a3, q2, a2, q1, a1, msg,
//		)
//	}
//	return
//}

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
