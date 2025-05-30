package models

import (
	"time"
)

type EmergencyContact struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Phone     string    `gorm:"unique;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
