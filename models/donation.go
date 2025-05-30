package models

import (
	"time"

	"gorm.io/gorm"
)

// Donation represents the donations table in the database.
// Fields correspond to kolom pada tabel donations.
type Donation struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null;column:title" json:"title"`
	Description string         `gorm:"type:text;column:description" json:"description"`
	QRImage     string         `gorm:"type:varchar(255);column:qr_image" json:"qr_image"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

// TableName overrides the table name used by GORM.
func (Donation) TableName() string {
	return "donations"
}
