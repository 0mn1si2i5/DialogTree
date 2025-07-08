// Path: ./models/image_model.go

package models

type ImageModel struct {
	Model
	Filename string `gorm:"size:64; not null" json:"filename"`
	Path     string `gorm:"size:256" json:"path"`
	Url      string `gorm:"size:256" json:"url"`
	Size     int64  `gorm:"not null" json:"size"`
	Hash     string `gorm:"size:64; not null; unique" json:"hash"`
	Source   string `gorm:"size:256" json:"source"`
}
