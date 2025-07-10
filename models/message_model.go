// Path: ./models/message_model.go

package models

type MessageModel struct {
	Model
	Prompt    string `json:"prompt"`
	Answer    string `json:"answer"`
	DialogID  uint   `gorm:"index" json:"dialogID"`
	ParentID  *uint  `gorm:"index" json:"parentID"`          // 不一定有
	IsStarred bool   `gorm:"default:false" json:"isStarred"` // 标星
	Comment   string `json:"comment"`                        // 评论
	Depth     int    `json:"depth"`
	Title     string `gorm:"size:64" json:"title"` // ai 归纳，给用户看
	Summary   string `json:"summary"`              // ai 归纳，维护上下文

	// fk
	DialogModel   DialogModel     `gorm:"foreignKey:DialogID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	ParentModel   *MessageModel   `gorm:"foreignKey:ParentID;references:ID" json:"-"`
	ChildrenModel []*MessageModel `gorm:"foreignKey:ParentID" json:"children"`
}
