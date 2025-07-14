// Path: ./models/conversation_model.go

package models

type ConversationModel struct {
	Model
	Prompt    string `json:"prompt"`
	Answer    string `json:"answer"`
	SessionID int64  `gorm:"index" json:"sessionID"`
	DialogID  int64  `gorm:"index" json:"dialogID"`
	IsStarred bool   `gorm:"default:false" json:"isStarred"` // 标星
	Comment   string `json:"comment"`                        // 评论
	Title     string `gorm:"size:64" json:"title"`           // ai 归纳，给用户看
	Summary   string `json:"summary"`                        // ai 归纳，维护上下文

	// fk
	SessionModel SessionModel `gorm:"foreignKey:SessionID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	DialogModel  DialogModel  `gorm:"foreignKey:DialogID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}
