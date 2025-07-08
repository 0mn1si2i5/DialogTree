// Path: ./models/Message_model.go

package models

type MessageModel struct {
	Model
	Prompt    string `json:"prompt"`
	Answer    string `json:"answer"`
	Abstract  string `gorm:"size:256" json:"abstract"` // ai生成的摘要
	DialogID  uint   `gorm:"index" json:"dialogID"`
	ParentID  *uint  `gorm:"index" json:"parentID"` // 不一定有
	IsStarred bool   `json:"isStarred"`             // 标星
	Comment   string `json:"comment"`               // 评论
	Depth     int    `json:"depth"`

	// fk
	DialogModel   DialogModel     `gorm:"foreignKey:DialogID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	ParentModel   *MessageModel   `gorm:"foreignKey:ParentID;references:ID" json:"-"`
	ChildrenModel []*MessageModel `gorm:"foreignKey:ParentID" json:"children"`
}
