// Path: ./models/dialog_model.go

package models

type DialogModel struct {
	Model
	SessionID int64  `gorm:"index"`
	ParentID  *int64 `gorm:"index"` // 对话之间的树状关系

	// fk
	SessionModel       SessionModel         `gorm:"foreignKey:SessionID;references:ID" json:"-"`
	ParentModel        *DialogModel         `gorm:"foreignKey:ParentID;references:ID" json:"-"`
	ChildrenModels     []*DialogModel       `gorm:"foreignKey:ParentID;references:ID" json:"-"`
	ConversationModels []*ConversationModel `gorm:"foreignKey:DialogID;references:ID" json:"-"`
}
