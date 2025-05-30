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

func CreateKorban(c *fiber.Ctx) error {
	var korban models.Korban
	if err := c.BodyParser(&korban); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	korban.NoRegistrasi = c.FormValue("no_registrasi")
	korban.NIKKorban = c.FormValue("nik_korban")
	korban.Nama = c.FormValue("nama_korban")
	usia, err := strconv.Atoi(c.FormValue("usia_korban"))
	if err == nil {
		korban.Usia = usia
	}
	korban.AlamatKorban = c.FormValue("alamat_korban")
	korban.AlamatDetail = c.FormValue("alamat_detail")
	korban.JenisKelamin = c.FormValue("jenis_kelamin")
	korban.Agama = c.FormValue("agama")
	korban.NoTelepon = c.FormValue("no_telepon")
	korban.Pendidikan = c.FormValue("pendidikan")
	korban.Pekerjaan = c.FormValue("pekerjaan")
	korban.StatusPerkawinan = c.FormValue("status_perkawinan")
	korban.Kebangsaan = c.FormValue("kebangsaan")
	korban.HubunganDenganKorban = c.FormValue("hubungan_dengan_pelaku")
	korban.KeteranganLainnya = c.FormValue("keterangan_lainnya")

	file, err := c.FormFile("dokumentasi_korban")
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
				Message: "Gagal Mengupload gambar",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}

		korban.DokumentasiPelaku = imageURL
	}
	korban.CreatedAt = time.Now()
	korban.UpdatedAt = time.Now()
	if err := database.DB.Create(&korban).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal Menambah Data Korban",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	response := helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Berhasil Menambah Data Korban",
		Data:    korban,
	}
	return c.Status(http.StatusCreated).JSON(response)
}

func UpdateKorban(c *fiber.Ctx) error {
	id := c.Params("id")
	var korban models.Korban
	if err := database.DB.First(&korban, id).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "korban not found",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}
	if err := c.BodyParser(&korban); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	if value := c.FormValue("no_registrasi"); value != "" {
		korban.NoRegistrasi = value
	}
	if value := c.FormValue("nik_korban"); value != "" {
		korban.NIKKorban = value
	}
	if value := c.FormValue("nama_korban"); value != "" {
		korban.Nama = value
	}
	if value := c.FormValue("usia_korban"); value != "" {
		if usia, err := strconv.Atoi(value); err == nil {
			korban.Usia = usia
		}
	}
	if value := c.FormValue("alamat_korban"); value != "" {
		korban.AlamatKorban = value
	}
	if value := c.FormValue("alamat_detail"); value != "" {
		korban.AlamatDetail = value
	}
	if value := c.FormValue("jenis_kelamin"); value != "" {
		korban.JenisKelamin = value
	}
	if value := c.FormValue("agama"); value != "" {
		korban.Agama = value
	}
	if value := c.FormValue("no_telepon"); value != "" {
		korban.NoTelepon = value
	}
	if value := c.FormValue("pendidikan"); value != "" {
		korban.Pendidikan = value
	}
	if value := c.FormValue("pekerjaan"); value != "" {
		korban.Pekerjaan = value
	}
	if value := c.FormValue("status_perkawinan"); value != "" {
		korban.StatusPerkawinan = value
	}
	if value := c.FormValue("kebangsaan"); value != "" {
		korban.Kebangsaan = value
	}
	if value := c.FormValue("hubungan_dengan_korban"); value != "" {
		korban.HubunganDenganKorban = value
	}
	if value := c.FormValue("keterangan_lainnya"); value != "" {
		korban.KeteranganLainnya = value
	}

	file, err := c.FormFile("dokumentasi_korban")
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
				Message: "Failed to upload image",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}

		korban.DokumentasiPelaku = imageURL
	}

	korban.UpdatedAt = time.Now()
	if err := database.DB.Save(&korban).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to update korban",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Pelaku updated successfully",
		Data:    korban,
	}
	return c.Status(http.StatusOK).JSON(response)
}
