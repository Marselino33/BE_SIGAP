package models

import "time"

type ViolenceCategory struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	CategoryName string    `json:"category_name"`
	Image        string    `json:"image"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
