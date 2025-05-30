package models

import (
	"time"

	"gorm.io/datatypes"
)

type TrackingLaporan struct {
	ID           uint              `gorm:"primaryKey" json:"id"`
	NoRegistrasi string            `gorm:"not null" json:"no_registrasi"`
	Keterangan   string            `json:"keterangan"`
	Document     datatypes.JSONMap `json:"document" form:"image" gorm:"type:json"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}
