// Path: ./models/category_model.go

package models

type CategoryModel struct {
	Model
	Name string `gorm:"not null;uniqueIndex:idx_uniq_category_name;size:32" json:"name"`

	// FK
	Sessions []SessionModel `gorm:"foreignKey:CategoryID;references:ID" json:"-"`
}
