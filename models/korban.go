package models

import (
	"time"
)

type Korban struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	NoRegistrasi         string    `json:"no_registrasi"`
	NIKKorban            string    `json:"nik_korban"`
	Nama                 string    `json:"nama_korban"`
	Usia                 int       `json:"usia_korban"`
	AlamatKorban         string    `json:"alamat_korban"`
	AlamatDetail         string    `json:"alamat_detail"`
	JenisKelamin         string    `json:"jenis_kelamin"`
	Agama                string    `json:"agama"`
	NoTelepon            string    `json:"no_telepon"`
	Pendidikan           string    `json:"pendidikan"`
	Pekerjaan            string    `json:"pekerjaan"`
	StatusPerkawinan     string    `json:"status_perkawinan"`
	Kebangsaan           string    `json:"kebangsaan"`
	HubunganDenganKorban string    `json:"hubungan_dengan_pelaku"`
	KeteranganLainnya    string    `json:"keterangan_lainnya"`
	DokumentasiPelaku    string    `json:"dokumentasi_korban"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
