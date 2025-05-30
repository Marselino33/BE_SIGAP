package models

import (
	"time"

	"gorm.io/gorm"
)

type ReportAdmin struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement"`
	ReportID       string         `gorm:"type:varchar(50);unique;not null"`
	ReporterID     int64          `gorm:"not null"`
	ReportedUserID uint64         `gorm:"not null"`
	ChatMessageID  string         `gorm:"type:varchar(50);not null"`
	MessageContent string         `gorm:"type:text;not null"`
	ReportType     string         `gorm:"type:varchar(20);not null"` // client_report, admin_notification
	Status         string         `gorm:"type:varchar(20);default:pending"`
	Notes          string         `gorm:"type:text"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"` // Menggunakan autoCreateTime
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"` // Menggunakan autoUpdateTime
	UpdatedBy      *int64         `gorm:""`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
