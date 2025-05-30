package models

import (
	"time"

	"gorm.io/datatypes"
)

type Laporan struct {
	NoRegistrasi        string            `gorm:"primaryKey" json:"no_registrasi"`
	User                User              `gorm:"foreignKey:UserID"`
	UserID              uint              `json:"user_id"`
	ViolenceCategory    ViolenceCategory  `gorm:"foreignKey:KategoriKekerasanID"`
	KategoriKekerasanID uint              `json:"kategori_kekerasan_id"`
	TanggalPelaporan    time.Time         `json:"tanggal_pelaporan"`
	TanggalKejadian     time.Time         `json:"tanggal_kejadian"`
	KategoriLokasiKasus string            `json:"kategori_lokasi_kasus"`
	AlamatTKP           string            `json:"alamat_tkp"`
	AlamatDetailTKP     string            `json:"alamat_detail_tkp"`
	KronologisKasus     string            `json:"kronologis_kasus"`
	Status              string            `json:"status"`
	AlasanDibatalkan    string            `json:"alasan_dibatalkan"`
	WaktuDilihat        *time.Time        `json:"waktu_dilihat"`
	UserIDMelihat       *uint             `json:"userid_melihat,omitempty"`
	WaktuDiproses       *time.Time        `json:"waktu_diproses"`
	WaktuDibatalkan     *time.Time        `json:"waktu_dibatalkan"`
	Dokumentasi         datatypes.JSONMap `json:"dokumentasi" form:"image" gorm:"type:json"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}
