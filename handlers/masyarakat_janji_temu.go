package handlers

import (
	"backend-pedika-fiber/auth"
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func MasyarakatCreateJanjiTemu(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	userID, err := auth.ExtractUserIDFromToken(token)
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized",
		}
		return c.Status(http.StatusUnauthorized).JSON(response)
	}

	var janjitemu models.JanjiTemu
	if err := c.BodyParser(&janjitemu); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	waktuDimulai, err := time.Parse("2006-01-02T15:04:05", c.FormValue("waktu_dimulai"))
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid format for start time",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	waktuSelesai, err := time.Parse("2006-01-02T15:04:05", c.FormValue("waktu_selesai"))
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid format for end time",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	janjitemu.WaktuDimulai = waktuDimulai
	janjitemu.WaktuSelesai = waktuSelesai
	janjitemu.Status = "Belum disetujui"
	janjitemu.KeperluanKonsultasi = c.FormValue("keperluan_konsultasi")
	janjitemu.UserID = uint(userID)
	janjitemu.UserIDTolakSetujui = nil

	if err := database.DB.Create(&janjitemu).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to create janjitemu",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	responseData := struct {
		ID                  uint      `json:"id"`
		UserID              uint      `json:"user_id"`
		WaktuDimulai        time.Time `json:"waktu_dimulai"`
		WaktuSelesai        time.Time `json:"waktu_selesai"`
		KeperluanKonsultasi string    `json:"keperluan_konsultasi"`
		Status              string    `json:"status"`
		UserTolakSetujui    uint      `json:"user_tolak_setujui"`
		AlasanDitolak       string    `json:"alasan_ditolak"`
		AlasanDibatalkan    string    `json:"alasan_dibatalkan"`
	}{
		ID:                  janjitemu.ID,
		UserID:              janjitemu.UserID,
		WaktuDimulai:        janjitemu.WaktuDimulai,
		WaktuSelesai:        janjitemu.WaktuSelesai,
		KeperluanKonsultasi: janjitemu.KeperluanKonsultasi,
		Status:              janjitemu.Status,
		UserTolakSetujui:    0,
		AlasanDitolak:       janjitemu.AlasanDitolak,
		AlasanDibatalkan:    janjitemu.AlasanDibatalkan,
	}

	response := helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Janjitemu created successfully",
		Data:    responseData,
	}
	return c.Status(http.StatusCreated).JSON(response)
}

func MasyarakatEditJanjiTemu(c *fiber.Ctx) error {
	janjiTemuID := c.Params("id")

	// Struct untuk menerima input dari client
	var updateRequest struct {
		WaktuDimulaiStr     string `json:"waktu_dimulai"`
		WaktuSelesaiStr     string `json:"waktu_selesai"`
		KeperluanKonsultasi string `json:"keperluan_konsultasi"`
	}

	// Parse body request
	if err := c.BodyParser(&updateRequest); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Body request tidak valid",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	// Log input untuk debugging
	log.Printf("Input diterima: waktu_dimulai=%s, waktu_selesai=%s, keperluan=%s",
		updateRequest.WaktuDimulaiStr, updateRequest.WaktuSelesaiStr, updateRequest.KeperluanKonsultasi)

	// Fungsi untuk parsing datetime dengan beberapa format
	parseDateTime := func(s string) (time.Time, error) {
		// Daftar format yang mungkin
		formats := []string{
			"2006-01-02T15:04:05.000Z07:00", // Dengan zona waktu (misalnya, 2025-05-15T04:57:00.000+07:00)
			"2006-01-02T15:04:05.000",       // Tanpa zona waktu (misalnya, 2025-05-15T04:57:00.000)
			"2006-01-02T15:04:05Z07:00",     // Tanpa milidetik, dengan zona waktu
			"2006-01-02T15:04:05",           // Tanpa milidetik dan zona waktu
			time.RFC3339,                    // Format RFC3339 standar
		}

		for _, format := range formats {
			if t, err := time.Parse(format, s); err == nil && !t.IsZero() {
				// Jika tidak ada zona waktu, anggap sebagai UTC
				if t.Location() == time.UTC || t.Location() == nil {
					return t.UTC(), nil
				}
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("format datetime tidak valid: %s", s)
	}

	// Parsing waktu_dimulai
	waktuDimulai, err := parseDateTime(updateRequest.WaktuDimulaiStr)
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Format waktu_dimulai tidak valid: " + err.Error(),
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	// Parsing waktu_selesai
	waktuSelesai, err := parseDateTime(updateRequest.WaktuSelesaiStr)
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Format waktu_selesai tidak valid: " + err.Error(),
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	// Validasi waktu tidak boleh nol
	if waktuDimulai.IsZero() || waktuSelesai.IsZero() {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Waktu dimulai atau waktu selesai tidak boleh kosong",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	// Validasi logis
	if waktuSelesai.Before(waktuDimulai) {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Waktu selesai harus setelah waktu dimulai",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	// Cari data janji temu
	var janjiTemu models.JanjiTemu
	if err := database.DB.First(&janjiTemu, janjiTemuID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Janji temu tidak ditemukan",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}

	// Cek status
	if janjiTemu.Status != "Belum disetujui" {
		response := helper.ResponseWithOutData{
			Code:    http.StatusForbidden,
			Status:  "error",
			Message: "Hanya janji temu dengan status 'Belum disetujui' yang bisa diedit",
		}
		return c.Status(http.StatusForbidden).JSON(response)
	}

	// Update data
	janjiTemu.WaktuDimulai = waktuDimulai
	janjiTemu.WaktuSelesai = waktuSelesai
	janjiTemu.KeperluanKonsultasi = updateRequest.KeperluanKonsultasi

	// Simpan ke database
	if err := database.DB.Save(&janjiTemu).Error; err != nil {
		log.Printf("Gagal menyimpan ke database: %v", err)
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal memperbarui janji temu: " + err.Error(),
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	// Respon sukses
	response := helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Janji temu berhasil diperbarui",
	}
	return c.Status(http.StatusOK).JSON(response)
}

func GetUserJanjiTemus(c *fiber.Ctx) error {
	userID, err := auth.ExtractUserIDFromToken(c.Get("Authorization"))
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized",
		}
		return c.Status(http.StatusUnauthorized).JSON(response)
	}
	var janjiTemus []models.JanjiTemu
	if err := database.DB.Preload("User").Preload("UserTolakSetujui").Where("user_id = ?", userID).Find(&janjiTemus).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to get user JanjiTemu records",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	if len(janjiTemus) == 0 {
		response := helper.ResponseWithOutData{
			Code:    http.StatusOK,
			Status:  "success",
			Message: "No JanjiTemu records found for the user",
		}
		return c.Status(http.StatusOK).JSON(response)
	}
	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "List of JanjiTemu by user",
		Data:    janjiTemus,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func GetJanjiTemuByID(c *fiber.Ctx) error {
	janjiTemuID := c.Params("id")
	var janjiTemu models.JanjiTemu
	if err := database.DB.Preload("UserTolakSetujui").First(&janjiTemu, janjiTemuID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "JanjiTemu not found",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}
	if janjiTemu.Status == "Ditolak" && janjiTemu.UserTolakSetujui.ID != 0 {
		var user models.User
		if err := database.DB.First(&user, janjiTemu.UserIDTolakSetujui).Error; err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Failed to fetch user detail who rejected the appointment",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		janjiTemu.UserTolakSetujui = user
	}
	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "JanjiTemu detail",
		Data:    janjiTemu,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func MasyarakatCancelJanjiTemu(c *fiber.Ctx) error {
	janjiTemuID := c.Params("id")

	var janjiTemu models.JanjiTemu
	if err := database.DB.First(&janjiTemu, janjiTemuID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Janji temu not found",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}
	if janjiTemu.Status != "Belum disetujui" {
		response := helper.ResponseWithOutData{
			Code:    http.StatusForbidden,
			Status:  "error",
			Message: "Forbidden: You can only cancel appointments with status 'Belum disetujui'",
		}
		return c.Status(http.StatusForbidden).JSON(response)
	}
	janjiTemu.Status = "Dibatalkan"
	janjiTemu.AlasanDibatalkan = c.FormValue("alasan_dibatalkan")

	if err := database.DB.Save(&janjiTemu).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to cancel janji temu",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Janji temu canceled successfully",
	}
	return c.Status(http.StatusOK).JSON(response)
}

func AdminGetAllJanjiTemu(c *fiber.Ctx) error {
	var janjiTemus []models.JanjiTemu
	if err := database.DB.Preload("User").Preload("UserTolakSetujui").Find(&janjiTemus).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to retrieve Janji Temu data",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "List of Janji Temu",
		Data:    janjiTemus,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func AdminJanjiTemuByID(c *fiber.Ctx) error {
	janjiTemuID := c.Params("id")
	var janjiTemu models.JanjiTemu
	if err := database.DB.Preload("UserTolakSetujui").Preload("User").First(&janjiTemu, janjiTemuID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "JanjiTemu not found",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}
	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "JanjiTemu detail",
		Data:    janjiTemu,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func AdminApproveJanjiTemu(c *fiber.Ctx) error {
	// Ambil token dari header Authorization
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: 0,
			Message: "Internal server error: Unable to retrieve user token",
			Data:    nil,
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: 0,
			Message: "Internal server error: Invalid token claims",
			Data:    nil,
		})
	}

	// Ambil userID dari claims
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: 0,
			Message: "Internal server error: Invalid user ID in token",
			Data:    nil,
		})
	}
	userID := uint(userIDFloat)

	// Ambil ID janji temu dari parameter URL
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Janji temu ID is required",
		})
	}

	// Cari janji temu di database
	var janjiTemu models.JanjiTemu
	db := database.GetGormDBInstance()
	if err := db.First(&janjiTemu, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Janji temu tidak ditemukan",
		})
	}

	// Update status janji temu
	janjiTemu.UserIDTolakSetujui = &userID // ID admin yang menyetujui
	janjiTemu.Status = "Disetujui"
	now := time.Now()
	if err := db.Save(&janjiTemu).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal menyimpan perubahan status",
		})
	}

	// Cari pengguna untuk notifikasi
	var user models.User
	if err := db.Where("id = ?", janjiTemu.UserID).First(&user).Error; err != nil {
		log.Printf("Failed to retrieve user for notification: %v", err)
	} else if user.NotificationToken != "" {
		log.Println("User notification token:", user.NotificationToken)

		// Siapkan data notifikasi FCM
		notificationData := models.FCMNotificationData{
			Type:      "appointment",
			ReportID:  id, // Gunakan ID janji temu sebagai identifier
			Status:    "approved",
			UpdatedBy: userID, // ID admin yang menyetujui
			UpdatedAt: now.Format(time.RFC3339),
			Notes:     "Kami sudah siap bertemu dengan Anda!",
			DeepLink:  "laporanku://appointments/" + id, // Deep link opsional
		}

		// Buat notifikasi untuk disimpan di database dengan pesan interaktif

		notification, err := NewNotificationFromFCMData(
			janjiTemu.UserID,
			"Yay! Janji Pertemuan Kamu Telah Disetujui!",
			"Hore! Jadwal janji temu kamu pada "+janjiTemu.WaktuDimulai.Format(time.RFC3339)+" telah disetujui. Jangan lupakan janji kita ya!",
			notificationData,
			now,
		)
		if err != nil {
			log.Printf("Error creating notification: %v", err)
		} else {

			if err := db.Create(&notification).Error; err != nil {
				log.Printf("Failed to store notification: %v", err)
			}

			if err := SendFCMNotification(user.NotificationToken, notificationData, *notification); err != nil {
				log.Printf("Failed to send FCM notification: %v", err)
			}
		}
	}

	// Response sukses
	response := helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Janji Temu berhasil disetujui",
	}
	return c.Status(http.StatusOK).JSON(response)
}

func AdminCancelJanjiTemu(c *fiber.Ctx) error {
	janjiTemuID := c.Params("id")
	if janjiTemuID == "" {
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Janji temu ID is required",
		})
	}

	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: 0,
			Message: "Internal server error: Unable to retrieve user token",
			Data:    nil,
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: 0,
			Message: "Internal server error: Invalid token claims",
			Data:    nil,
		})
	}

	// Ambil userID dari claims
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: 0,
			Message: "Internal server error: Invalid user ID in token",
			Data:    nil,
		})
	}
	userID := uint(userIDFloat)
	log.Println("[DEBUG] Content-Type:", c.Get("Content-Type"))
	log.Println("[DEBUG] Raw Body:", string(c.Body()))

	// Parse body untuk alasan ditolak
	alasanDitolak := c.FormValue("alasan_ditolak")
	if alasanDitolak == "" {
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Alasan ditolak is required",
		})
	}

	// Cari janji temu di database
	var janjiTemu models.JanjiTemu
	db := database.GetGormDBInstance()
	if err := db.First(&janjiTemu, janjiTemuID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Janji temu not found",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}

	// Update status janji temu
	janjiTemu.Status = "Ditolak"
	janjiTemu.UserIDTolakSetujui = &userID
	janjiTemu.AlasanDitolak = alasanDitolak
	now := time.Now()
	if err := db.Save(&janjiTemu).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to cancel janji temu",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	// Cari pengguna untuk notifikasi
	var user models.User
	if err := db.Where("id = ?", janjiTemu.UserID).First(&user).Error; err != nil {
		log.Printf("Failed to retrieve user for notification: %v", err)
	} else if user.NotificationToken != "" {
		log.Println("User notification token:", user.NotificationToken)

		notificationData := models.FCMNotificationData{
			Type:      "appointment",
			ReportID:  janjiTemuID,
			Status:    "rejected",
			UpdatedBy: userID,
			UpdatedAt: now.Format(time.RFC3339),
			Notes:     "Maaf, janji temu Anda ditolak karena: " + alasanDitolak,
			DeepLink:  "laporanku://appointments/" + janjiTemuID,
		}

		notification, err := NewNotificationFromFCMData(
			janjiTemu.UserID,
			"Oops! Janji Pertemuan Kamu Ditolak..",
			"Sayang sekali, janji temu kamu pada "+janjiTemu.WaktuDimulai.Format(time.RFC3339)+" ditolak. Alasan: "+janjiTemu.AlasanDitolak,
			notificationData,
			now,
		)
		if err != nil {
			log.Printf("Error creating notification: %v", err)
		} else {
			// Simpan notifikasi ke database
			if err := db.Create(notification).Error; err != nil {
				log.Printf("Failed to store notification: %v", err)
			}

			// Kirim push notification via FCM
			if err := SendFCMNotification(user.NotificationToken, notificationData, *notification); err != nil {
				log.Printf("Failed to send FCM notification: %v", err)
			}
		}
	}

	// Response sukses
	response := helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Janji Temu Sudah Ditolak",
	}
	return c.Status(http.StatusOK).JSON(response)
}
