package models

import (
	"time"
)

type Event struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	NamaEvent          string    `json:"nama_event"`
	DeskripsiEvent     string    `json:"deskripsi_event"`
	ThumbnailEvent     string    `json:"thumbnail_event"`
	TanggalPelaksanaan time.Time `json:"tanggal_pelaksanaan"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
