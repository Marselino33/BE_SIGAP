package models

import "time"

// Notification merepresentasikan data notifikasi pengguna
type Notification struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"` // Foreign key ke users
    Type      string    `gorm:"not null" json:"type"`          // report_status, chat, appointment, INIII BARU 3 KALO MAU DITAMBAH JAN LUPA HANDLE FRONT END JUGA NOTIFICATIONS SCREEN.
    Title     string    `gorm:"not null" json:"title"`
    Body      string    `gorm:"not null" json:"body"`
    Data      string    `gorm:"type:json" json:"data"`         // JSON dari Stucut FCMNotificationData
    IsRead    bool      `gorm:"default:false" json:"is_read"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type FCMNotificationData struct {
	Type      string `json:"type"` // report_status, chat, appointment, INIII BARU 3 KALO MAU DITAMBAH JAN LUPA HANDLE FRONT END JUGA NOTIFICATIONS CHANNEL/Notification Service.
	ReportID  string `json:"reportId"`
	Status    string `json:"status"`
	UpdatedBy uint   `json:"updatedBy"`
	UpdatedAt string `json:"updatedAt"`
	Notes     string `json:"notes"`
	DeepLink  string `json:"deepLink"`
	ImageURL  string `json:"imageUrl,omitempty"`
}

type FCMNotificationContent struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type FCMMessage struct {
	Token        string                 `json:"token"`
	Data         FCMNotificationData    `json:"data"`
	Notification FCMNotificationContent `json:"notification"`
}

