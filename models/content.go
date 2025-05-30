package models

import (
	"time"
)

type Content struct {
	ID                 uint             `json:"id" gorm:"primaryKey"`
	Judul              string           `json:"judul"`
	IsiContent         string           `json:"isi_content"`
	ImageContent       string           `json:"image_content"`
	ViolenceCategory   ViolenceCategory `json:"violence_category" gorm:"foreignKey:ViolenceCategoryID"`
	ViolenceCategoryID uint             `json:"violence_category_id"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}
