package models

import "time"

type JanjiTemu struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	User                User      `json:"user" gorm:"foreignKey:UserID"`
	UserID              uint      `json:"user_id"`
	WaktuDimulai        time.Time `json:"waktu_dimulai"`
	WaktuSelesai        time.Time `json:"waktu_selesai"`
	KeperluanKonsultasi string    `json:"keperluan_konsultasi"`
	Status              string    `json:"status"`
	UserTolakSetujui    User      `json:"user_tolak_setujui" gorm:"foreignKey:UserIDTolakSetujui"`
	UserIDTolakSetujui  *uint     `json:"userid_tolak_setujui,omitempty"`
	AlasanDitolak       string    `json:"alasan_ditolak" gorm:"column:alasan_ditolak"`
	AlasanDibatalkan    string    `json:"alasan_dibatalkan" gorm:"column:alasan_dibatalkan"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
