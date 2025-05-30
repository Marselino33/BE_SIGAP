package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreatePelaku(c *fiber.Ctx) error {
	var pelaku models.Pelaku
	// if err := c.BodyParser(&pelaku); err != nil {
	// 	response := helper.ResponseWithOutData{
	// 		Code:    http.StatusBadRequest,
	// 		Status:  "error",
	// 		Message: "Invalid request body",
	// 	}
	// 	return c.Status(http.StatusBadRequest).JSON(response)
	// }
	pelaku.NoRegistrasi = c.FormValue("no_registrasi")
	pelaku.NIKPelaku = c.FormValue("nik_pelaku")
	pelaku.Nama = c.FormValue("nama_pelaku")
	usia, err := strconv.Atoi(c.FormValue("usia_pelaku"))
	if err == nil {
		pelaku.Usia = usia
	}
	pelaku.AlamatPelaku = c.FormValue("alamat_pelaku")
	pelaku.AlamatDetail = c.FormValue("alamat_detail")
	pelaku.JenisKelamin = c.FormValue("jenis_kelamin")
	pelaku.Agama = c.FormValue("agama")
	pelaku.NoTelepon = c.FormValue("no_telepon")
	pelaku.Pendidikan = c.FormValue("pendidikan")
	pelaku.Pekerjaan = c.FormValue("pekerjaan")
	pelaku.StatusPerkawinan = c.FormValue("status_perkawinan")
	pelaku.Kebangsaan = c.FormValue("kebangsaan")
	pelaku.HubunganDenganKorban = c.FormValue("hubungan_dengan_korban")
	pelaku.KeteranganLainnya = c.FormValue("keterangan_lainnya")

	file, err := c.FormFile("dokumentasi_pelaku")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Failed to open image file",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		defer src.Close()

		imageURL, err := helper.UploadFileToCloudinary(src, file.Filename)
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal Mengupload Gambar",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}

		pelaku.DokumentasiPelaku = imageURL
	}
	pelaku.CreatedAt = time.Now()
	pelaku.UpdatedAt = time.Now()
	if err := database.DB.Create(&pelaku).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "gagal Menambahkan Data pelaku",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	response := helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Berhasil Menambah Data Pelaku",
		Data:    pelaku,
	}
	return c.Status(http.StatusCreated).JSON(response)
}

func UpdatePelaku(c *fiber.Ctx) error {
	id := c.Params("id")
	var pelaku models.Pelaku

	// 1) Cari dulu record
	if err := database.DB.First(&pelaku, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Pelaku Tidak Ditemukan",
		})
	}

	// 2) Ambil semua field via FormValue, TIDAK pakai BodyParser
	if v := c.FormValue("no_registrasi"); v != "" {
		pelaku.NoRegistrasi = v
	}
	if v := c.FormValue("nik_pelaku"); v != "" {
		pelaku.NIKPelaku = v
	}
	if v := c.FormValue("nama_pelaku"); v != "" {
		pelaku.Nama = v
	}
	if v := c.FormValue("usia_pelaku"); v != "" {
		if usia, err := strconv.Atoi(v); err == nil {
			pelaku.Usia = usia
		}
	}
	if v := c.FormValue("alamat_pelaku"); v != "" {
		pelaku.AlamatPelaku = v
	}
	if v := c.FormValue("alamat_detail"); v != "" {
		pelaku.AlamatDetail = v
	}
	if v := c.FormValue("jenis_kelamin"); v != "" {
		pelaku.JenisKelamin = v
	}
	if v := c.FormValue("agama"); v != "" {
		pelaku.Agama = v
	}
	if v := c.FormValue("no_telepon"); v != "" {
		pelaku.NoTelepon = v
	}
	if v := c.FormValue("pendidikan"); v != "" {
		pelaku.Pendidikan = v
	}
	if v := c.FormValue("pekerjaan"); v != "" {
		pelaku.Pekerjaan = v
	}
	if v := c.FormValue("status_perkawinan"); v != "" {
		pelaku.StatusPerkawinan = v
	}
	if v := c.FormValue("kebangsaan"); v != "" {
		pelaku.Kebangsaan = v
	}
	if v := c.FormValue("hubungan_dengan_korban"); v != "" {
		pelaku.HubunganDenganKorban = v
	}
	if v := c.FormValue("keterangan_lainnya"); v != "" {
		pelaku.KeteranganLainnya = v
	}

	// 3) Handle file upload (jika ada)
	if file, err := c.FormFile("dokumentasi_pelaku"); err == nil {
		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Failed to open image file",
			})
		}
		defer src.Close()

		imageURL, err := helper.UploadFileToCloudinary(src, file.Filename)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal Mengupload Gambar",
			})
		}
		pelaku.DokumentasiPelaku = imageURL
	}

	pelaku.UpdatedAt = time.Now()

	// 4) Simpan perubahan—pakai Save atau Updates
	if err := database.DB.Save(&pelaku).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal Mengupdate Data Pelaku",
		})
	}

	// 5) Kembalikan response sukses
	return c.Status(fiber.StatusOK).JSON(helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Berhasil Mengupdate Data Pelaku",
		Data:    pelaku,
	})
}

func DeletePelaku(c *fiber.Ctx) error {
	id := c.Params("id")
	var pelaku models.Pelaku
	if err := database.DB.First(&pelaku, id).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Pelaku Tidak Ditemukan",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}

	if err := database.DB.Delete(&pelaku).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal Menghapus Data Pelaku",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Berhasil Menghapus Data Pelaku",
	}
	return c.Status(http.StatusOK).JSON(response)
}
