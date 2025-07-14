// Path: ./models/session_model.go

package models

type SessionModel struct {
	Model
	Tittle       string `gorm:"size:64" json:"tittle"`
	Summary      string `gorm:"size:256" json:"summary"`
	CategoryID   int64  `json:"categoryID"`
	RootDialogID *int64 `json:"rootDialogID"`

	// fk
	RootDialogModel *DialogModel   `gorm:"foreignKey:RootDialogID;references:ID" json:"-"`
	CategoryModel   *CategoryModel `gorm:"foreignKey:CategoryID;references:ID" json:"-"`
}
