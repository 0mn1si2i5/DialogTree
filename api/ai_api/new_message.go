// Path: ./api/ai_api/new_message.go

package ai_api

import (
	"dialogTree/common/res"
	"dialogTree/global"
	"dialogTree/models"
	"dialogTree/service/ai_service/chat_anywhere_ai"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"strings"
)

type AiChatReq struct {
	Content    string `json:"content" binding:"required"`
	DialogID   uint   `json:"dialogID"`
	ParentID   *uint  `json:"parentID"`
	CategoryID uint   `json:"categoryID"`
}

func (AiApi) NewMessageView(c *gin.Context) {
	req := c.MustGet("req").(AiChatReq)

	var msg string
	if req.ParentID != nil {
		var mmodel models.MessageModel
		global.DB.Find(&mmodel, *req.ParentID)
		prevSum := mmodel.Summary
		msg += fmt.Sprintf("上轮对话：%s;本轮问题：", prevSum)
	}
	msg += req.Content
	msgChan, err := chat_anywhere_ai.ChatStream(msg)
	if err != nil {
		res.Fail(err, "链接失败", c)
		return
	}
	var ans string
	for s := range msgChan {
		res.SSESuccess(s, c)
		ans += s
	}
	go Archive(msg, ans, req)
}

type summarizeType struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

func Archive(q, a string, req AiChatReq) {
	msg := fmt.Sprintf("userQuestion:%s;aiResponse:%s;", q, a)
	rep, err := chat_anywhere_ai.Summarize(msg)
	if err != nil {
		logrus.Errorf("summarize error: %v\n", err)
		return
	}
	rep = extractJSON(rep)

	var s summarizeType
	err = json.Unmarshal([]byte(rep), &s)
	if err != nil {
		logrus.Errorf("json unmarshal error: %v\n", err)
		return
	}

	if req.CategoryID == 0 {
		req.CategoryID = 1
	}

	var isNew bool
	if req.DialogID == 0 {
		var dialog = models.DialogModel{
			CategoryID: req.CategoryID,
		}
		err = global.DB.Create(&dialog).Error
		if err != nil {
			logrus.Errorf("create dialog error: %v\n", err)
			return
		}
		req.DialogID = dialog.ID
		isNew = true
	}
	message := models.MessageModel{
		Prompt:   q,
		Answer:   a,
		DialogID: req.DialogID,
		ParentID: req.ParentID,
		Title:    s.Title,
		Summary:  s.Summary,
	}
	err = global.DB.Create(&message).Error
	if err != nil {
		logrus.Errorf("create dialog error: %v\n", err)
		return
	}

	if isNew {
		update := map[string]interface{}{
			"abstract":        s.Summary,
			"root_message_id": message.ID,
		}
		err = global.DB.Model(&models.DialogModel{}).Where("id = ?", req.DialogID).Updates(update).Error
		if err != nil {
			logrus.Errorf("update dialog error: %v\n", err)
			return
		}
	}
	return
}

func extractJSON(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
	}
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
	}
	if strings.HasSuffix(content, "```") {
		content = strings.TrimSuffix(content, "```")
	}
	return strings.TrimSpace(content)
}
