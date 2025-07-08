// Path: ./models/dialog_model.go

package models

type DialogModel struct {
	Model
	Abstract      string `gorm:"size:256" json:"abstract"`
	CategoryID    uint   `json:"categoryID"`
	RootMessageID *uint  `gorm:"unique;not null" json:"rootMessageID"`

	// fk
	RootMessageModel *MessageModel  `gorm:"foreignKey:RootMessageID;references:ID" json:"-"`
	CategoryModel    *CategoryModel `gorm:"foreignKey:CategoryID;references:ID" json:"-"`
}
