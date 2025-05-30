package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/admin"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// func CreateTrackingLaporan(c *fiber.Ctx) error {
// 	userToken, ok := c.Locals("user").(*jwt.Token)
// 	if !ok || userToken == nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Internal server error: Unable to retrieve user token",
// 		})
// 	}
// 	claims, ok := userToken.Claims.(jwt.MapClaims)
// 	if !ok {
// 		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Internal server error: Invalid token claims",
// 		})
// 	}

// 	// Ambil userID dari claims
// 	userIDFloat, ok := claims["user_id"].(float64)
// 	if !ok {
// 		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Internal server error: Invalid user ID in token",
// 		})
// 	}
// 	userID := uint(userIDFloat)
// 	// Struktur awal JSON
// 	var request struct {
// 		NoRegistrasi string   `json:"no_registrasi"`
// 		Keterangan   string   `json:"keterangan"`
// 		Document     []string `json:"document"`
// 	}

// 	// Coba parse body JSON
// 	_ = c.BodyParser(&request)

// 	// Jika kosong, cek dari form
// 	if request.NoRegistrasi == "" {
// 		request.NoRegistrasi = c.FormValue("no_registrasi")
// 	}
// 	if request.Keterangan == "" {
// 		request.Keterangan = c.FormValue("keterangan")
// 	}

// 	// Coba ambil document dari form-data (hanya 1 file untuk sekarang)
// 	formFile, err := c.FormFile("document")
// 	if err == nil && formFile != nil {
// 		// Kamu bisa simpan file, atau sekadar simpan nama sebagai representasi
// 		request.Document = []string{formFile.Filename}
// 	}

// 	fmt.Println("Log final: ", request.NoRegistrasi, request.Keterangan, request.Document)

// 	// Validasi wajib
// 	if request.NoRegistrasi == "" {
// 		log.Println("No Registrasi is empty (from JSON or Form)")
// 		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusBadRequest,
// 			Status:  "error",
// 			Message: "No Registrasi is required",
// 		})
// 	}
// 	// Cek keberadaan laporan
// 	var existingLaporan models.Laporan
// 	if err := database.GetGormDBInstance().Where("no_registrasi = ?", request.NoRegistrasi).First(&existingLaporan).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			log.Printf("No Registrasi %s not found", request.NoRegistrasi)
// 			response := helper.ResponseWithOutData{
// 				Code:    http.StatusBadRequest,
// 				Status:  "error",
// 				Message: "No Registrasi not found in Laporan table",
// 			}
// 			return c.Status(http.StatusBadRequest).JSON(response)
// 		}
// 		log.Printf("Database error: %v", err)
// 		response := helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Database error",
// 		}
// 		return c.Status(http.StatusInternalServerError).JSON(response)
// 	}

// 	// Validasi dokumen (opsional)
// 	imageURLs := request.Document

// 	// Buat tracking laporan
// 	trackingLaporan := models.TrackingLaporan{
// 		NoRegistrasi: request.NoRegistrasi,
// 		Keterangan:   request.Keterangan,
// 		Document:     datatypes.JSONMap{"urls": imageURLs},
// 		CreatedAt:    time.Now(),
// 		UpdatedAt:    time.Now(),
// 	}

// 	// Simpan ke database
// 	if err := database.GetGormDBInstance().Create(&trackingLaporan).Error; err != nil {
// 		log.Printf("Failed to create tracking laporan: %v", err)
// 		response := helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Failed to create tracking laporan",
// 		}
// 		return c.Status(http.StatusInternalServerError).JSON(response)
// 	}

// 	// Kirim notifikasi FCM
// 	var user models.User
// 	db := database.GetGormDBInstance()
// 	if err := db.Where("id = ?", existingLaporan.UserID).First(&user).Error; err == nil && user.NotificationToken != "" {
// 		notificationData := models.FCMNotificationData{
// 			Type:      "tracking_update",
// 			ReportID:  request.NoRegistrasi,
// 			Status:    "new_tracking",
// 			UpdatedBy: userID,
// 			UpdatedAt: time.Now().Format(time.RFC3339),
// 			Notes:     request.Keterangan,
// 			DeepLink:  "laporanku://tracking/" + request.NoRegistrasi,
// 		}

// 		docMessage := "Ada update baru nih!"
// 		if len(imageURLs) > 0 {
// 			docMessage = "Ada dokumen baru (PDF/Image) yang diunggah untuk laporanmu! 📎"
// 		}

// 		notification, err := NewNotificationFromFCMData(
// 			existingLaporan.UserID,
// 			"Update Baru pada Laporanmu!",
// 			"Halo! Tracking laporan dengan No. "+request.NoRegistrasi+" telah ditambahkan. "+docMessage+" Cek sekarang yuk!",
// 			notificationData,
// 			time.Now(),
// 		)
// 		if err != nil {
// 			log.Printf("Error creating notification: %v", err)
// 		} else {
// 			if err := db.Create(&notification).Error; err != nil {
// 				log.Printf("Failed to store notification: %v", err)
// 			}
// 			if err := SendFCMNotification(user.NotificationToken, notificationData, *notification); err != nil {
// 				log.Printf("Failed to send FCM notification: %v", err)
// 			}
// 		}
// 	}

// 	// Buat respons
// 	response := helper.ResponseWithData{
// 		Code:    http.StatusCreated,
// 		Status:  "success",
// 		Message: "Tracking laporan created successfully",
// 		Data: fiber.Map{
// 			"id":            trackingLaporan.ID,
// 			"no_registrasi": trackingLaporan.NoRegistrasi,
// 			"keterangan":    trackingLaporan.Keterangan,
// 			"document": fiber.Map{
// 				"urls": imageURLs,
// 			},
// 			"created_at": trackingLaporan.CreatedAt,
// 			"updated_at": trackingLaporan.UpdatedAt,
// 		},
// 	}

//		log.Printf("Tracking laporan for %s created successfully by user %d", request.NoRegistrasi, userID)
//		return c.Status(http.StatusCreated).JSON(response)
//	}
func CreateTrackingLaporan(c *fiber.Ctx) error {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Internal server error: Unable to retrieve user token",
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Internal server error: Invalid token claims",
		})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Internal server error: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)

	// Struktur awal JSON
	var request struct {
		NoRegistrasi string   `json:"no_registrasi" form:"no_registrasi"`
		Keterangan   string   `json:"keterangan" form:"keterangan"`
		Document     []string `json:"document"`
	}

	// Parse body JSON atau form-data
	if err := c.BodyParser(&request); err != nil && err != fiber.ErrUnprocessableEntity {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Failed to parse request body",
		})
	}

	// Ambil files dari form-data
	form, err := c.MultipartForm()
	var imageURLs []string
	if err == nil && form.File["document"] != nil {
		// Upload files ke Cloudinary
		fmt.Println("Log: ", form.File["document"], "Mengupload ke Cloudinary")
		imageURLs, err = helper.UploadMultipleFileToCloudinary(form.File["document"])
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: fmt.Sprintf("Failed to upload files to Cloudinary: %v", err),
			})
		}
		request.Document = imageURLs
	} else if len(request.Document) > 0 {
		// Jika document berisi URL (dari mobile), gunakan langsung
		// Opsional: tambahkan validasi URL jika diperlukan
		imageURLs = request.Document
	}

	fmt.Println("Log final: ", request.NoRegistrasi, request.Keterangan, request.Document)

	// Validasi wajib
	if request.NoRegistrasi == "" {
		log.Println("No Registrasi is empty")
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "No Registrasi is required",
		})
	}

	// Cek keberadaan laporan
	var existingLaporan models.Laporan
	if err := database.GetGormDBInstance().Where("no_registrasi = ?", request.NoRegistrasi).First(&existingLaporan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("No Registrasi %s not found", request.NoRegistrasi)
			return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
				Code:    http.StatusBadRequest,
				Status:  "error",
				Message: "No Registrasi not found in Laporan table",
			})
		}
		log.Printf("Database error: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Database error",
		})
	}

	// Buat tracking laporan
	trackingLaporan := models.TrackingLaporan{
		NoRegistrasi: request.NoRegistrasi,
		Keterangan:   request.Keterangan,
		Document:     datatypes.JSONMap{"urls": imageURLs},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Simpan ke database
	if err := database.GetGormDBInstance().Create(&trackingLaporan).Error; err != nil {
		log.Printf("Failed to create tracking laporan: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to create tracking laporan",
		})
	}

	// Kirim notifikasi FCM
	var user models.User
	db := database.GetGormDBInstance()
	if err := db.Where("id = ?", existingLaporan.UserID).First(&user).Error; err == nil && user.NotificationToken != "" {
		notificationData := models.FCMNotificationData{
			Type:      "tracking_update",
			ReportID:  request.NoRegistrasi,
			Status:    "new_tracking",
			UpdatedBy: userID,
			UpdatedAt: time.Now().Format(time.RFC3339),
			Notes:     request.Keterangan,
			DeepLink:  "laporanku://tracking/" + request.NoRegistrasi,
		}

		docMessage := "Ada update baru nih!"
		if len(imageURLs) > 0 {
			docMessage = "Ada dokumen baru (PDF/Image) yang diunggah untuk laporanmu! 📎"
		}

		notification, err := NewNotificationFromFCMData(
			existingLaporan.UserID,
			"Update Baru pada Laporanmu!",
			"Halo! Tracking laporan dengan No. "+request.NoRegistrasi+" telah ditambahkan. "+docMessage+" Cek sekarang yuk!",
			notificationData,
			time.Now(),
		)
		if err != nil {
			log.Printf("Error creating notification: %v", err)
		} else {
			if err := db.Create(notification).Error; err != nil {
				log.Printf("Failed to store notification: %v", err)
			}
			if err := SendFCMNotification(user.NotificationToken, notificationData, *notification); err != nil {
				log.Printf("Failed to send FCM notification: %v", err)
			}
		}
	}

	// Buat respons
	response := helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Tracking laporan created successfully",
		Data: fiber.Map{
			"id":            trackingLaporan.ID,
			"no_registrasi": trackingLaporan.NoRegistrasi,
			"keterangan":    trackingLaporan.Keterangan,
			"document": fiber.Map{
				"urls": imageURLs,
			},
			"created_at": trackingLaporan.CreatedAt,
			"updated_at": trackingLaporan.UpdatedAt,
		},
	}

	log.Printf("Tracking laporan for %s created successfully by user %d", request.NoRegistrasi, userID)
	return c.Status(http.StatusCreated).JSON(response)
}
func UpdateTrackingLaporan(c *fiber.Ctx) error {
	log.Println("===> Mulai UpdateTrackingLaporan")

	// Validasi token dan user ID
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		log.Println("❌ Gagal mendapatkan user token dari context")
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Internal server error: Unable to retrieve user token",
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("❌ Token claims tidak valid")
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Internal server error: Invalid token claims",
		})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		log.Println("❌ User ID tidak valid dalam token claims")
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Internal server error: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)

	// Validasi parameter ID
	trackingLaporanID := c.Params("id")
	if trackingLaporanID == "" {
		log.Println("❌ Parameter ID kosong")
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "ID is required",
		})
	}

	// Cari tracking laporan
	db := database.GetGormDBInstance()
	var trackingLaporan models.TrackingLaporan
	if err := db.First(&trackingLaporan, trackingLaporanID).Error; err != nil {
		log.Printf("❌ Gagal menemukan tracking laporan: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(http.StatusNotFound).JSON(helper.ResponseWithOutData{
				Code:    http.StatusNotFound,
				Status:  "error",
				Message: "Tracking Laporan not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Database error",
		})
	}

	// Parse body JSON atau form-data
	var request struct {
		NoRegistrasi string   `json:"no_registrasi" form:"no_registrasi"`
		Keterangan   string   `json:"keterangan" form:"keterangan"`
		Document     []string `json:"document" form:"document"`
	}
	if err := c.BodyParser(&request); err != nil && err != fiber.ErrUnprocessableEntity {
		log.Printf("❌ Gagal parsing body request: %v", err)
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Failed to parse request body",
		})
	}

	// Ambil files dari form-data
	var imageURLs []string
	form, err := c.MultipartForm()
	if err == nil && form.File["document"] != nil {
		// Upload files ke Cloudinary
		log.Println("Log: ", form.File["document"], "Mengupload ke Cloudinary")
		imageURLs, err = helper.UploadMultipleFileToCloudinary(form.File["document"])
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: fmt.Sprintf("Failed to upload files to Cloudinary: %v", err),
			})
		}
	} else if len(request.Document) > 0 {
		// Jika document berisi URL (dari mobile), gunakan langsung
		// Opsional: tambahkan validasi URL jika diperlukan
		imageURLs = request.Document
	}

	// Jika ada dokumen baru, hapus dokumen lama dari Cloudinary
	if len(imageURLs) > 0 && trackingLaporan.Document != nil {
		cloudName := os.Getenv("CLOUD_NAME")
		apiKey := os.Getenv("API_KEY")
		apiSecret := os.Getenv("API_SECRET")
		if cloudName == "" || apiKey == "" || apiSecret == "" {
			log.Println("❌ Missing Cloudinary credentials")
			return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Missing Cloudinary credentials",
			})
		}

		cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
		if err != nil {
			log.Printf("❌ Failed to initialize Cloudinary: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Failed to initialize Cloudinary",
			})
		}

		// Ambil URL lama dari dokumen
		var publicIDs []string
		if urls, ok := trackingLaporan.Document["urls"].([]interface{}); ok {
			for _, url := range urls {
				urlStr, ok := url.(string)
				if !ok {
					continue
				}
				// Ekstrak PublicID dari URL
				parts := strings.Split(urlStr, "/")
				if len(parts) > 0 {
					publicID := strings.Split(parts[len(parts)-1], ".")[0] // Ambil nama file tanpa ekstensi
					publicIDs = append(publicIDs, publicID)
				}
			}
		}

		// Hapus file lama dari Cloudinary
		if len(publicIDs) > 0 {
			ctx := context.Background()
			_, err = cld.Admin.DeleteAssets(ctx, admin.DeleteAssetsParams{
				PublicIDs:    publicIDs,
				DeliveryType: "upload",
				AssetType:    "image",
			})
			if err != nil {
				log.Printf("❌ Failed to delete old images from Cloudinary: %v", err)
				// Lanjutkan meski gagal hapus, tetapi log error
			}
		}
	}

	// Update fields jika ada
	if request.NoRegistrasi != "" {
		var laporan models.Laporan
		if err := db.Where("no_registrasi = ?", request.NoRegistrasi).First(&laporan).Error; err != nil {
			log.Printf("❌ No Registrasi %s not found: %v", request.NoRegistrasi, err)
			return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
				Code:    http.StatusBadRequest,
				Status:  "error",
				Message: "No Registrasi not found in Laporan table",
			})
		}
		trackingLaporan.NoRegistrasi = request.NoRegistrasi
	}
	if request.Keterangan != "" {
		trackingLaporan.Keterangan = request.Keterangan
	}
	if len(imageURLs) > 0 {
		trackingLaporan.Document = datatypes.JSONMap{"urls": imageURLs}
	}

	// Set waktu update
	trackingLaporan.UpdatedAt = time.Now()

	// Simpan perubahan ke database
	if err := db.Save(&trackingLaporan).Error; err != nil {
		log.Printf("❌ Gagal menyimpan perubahan ke database: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to update tracking laporan",
		})
	}

	// Kirim notifikasi FCM
	var laporan models.Laporan
	if err := db.Where("no_registrasi = ?", trackingLaporan.NoRegistrasi).First(&laporan).Error; err == nil {
		var user models.User
		if err := db.Where("id = ?", laporan.UserID).First(&user).Error; err == nil && user.NotificationToken != "" {
			notificationData := models.FCMNotificationData{
				Type:      "tracking_update",
				ReportID:  trackingLaporan.NoRegistrasi,
				Status:    "updated_tracking",
				UpdatedBy: userID,
				UpdatedAt: trackingLaporan.UpdatedAt.Format(time.RFC3339),
				Notes:     trackingLaporan.Keterangan,
				DeepLink:  "laporanku://tracking/" + trackingLaporan.NoRegistrasi,
			}

			docMessage := "Ada perubahan terbaru pada tracking laporanmu!"
			if len(imageURLs) > 0 {
				docMessage = "Dokumen baru (PDF/Image) telah diperbarui untuk laporanmu! 📎"
			}

			notification, err := NewNotificationFromFCMData(
				laporan.UserID,
				"Tracking Laporanmu Diperbarui!",
				"Yay! Tracking untuk laporan No. "+trackingLaporan.NoRegistrasi+" telah diperbarui. "+docMessage+" Yuk cek detailnya!",
				notificationData,
				trackingLaporan.UpdatedAt,
			)
			if err != nil {
				log.Printf("❌ Error creating notification: %v", err)
			} else {
				if err := db.Create(notification).Error; err != nil {
					log.Printf("❌ Failed to store notification: %v", err)
				}
				if err := SendFCMNotification(user.NotificationToken, notificationData, *notification); err != nil {
					log.Printf("❌ Failed to send FCM notification: %v", err)
				}
			}
		}
	}

	// Buat respons
	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Tracking laporan updated successfully",
		Data: fiber.Map{
			"id":            trackingLaporan.ID,
			"no_registrasi": trackingLaporan.NoRegistrasi,
			"keterangan":    trackingLaporan.Keterangan,
			"document": fiber.Map{
				"urls": imageURLs,
			},
			"created_at": trackingLaporan.CreatedAt,
			"updated_at": trackingLaporan.UpdatedAt,
		},
	}
	log.Printf("✅ Tracking laporan %s updated successfully by user %d", trackingLaporan.NoRegistrasi, userID)
	return c.Status(http.StatusOK).JSON(response)
}

// func UpdateTrackingLaporan(c *fiber.Ctx) error {
// 	log.Println("===> Mulai UpdateTrackingLaporan")
// 	userToken, ok := c.Locals("user").(*jwt.Token)
// 	if !ok || userToken == nil {
// 		log.Println("❌ Gagal mendapatkan user token dari context")
// 		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Internal server error: Unable to retrieve user token",
// 		})
// 	}
// 	claims, ok := userToken.Claims.(jwt.MapClaims)
// 	if !ok {
// 		log.Println("❌ Token claims tidak valid")
// 		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Internal server error: Invalid token claims",
// 		})
// 	}

// 	// Ambil userID dari claims
// 	userIDFloat, ok := claims["user_id"].(float64)
// 	if !ok {
// 		log.Println("❌ User ID tidak valid dalam token claims")
// 		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Internal server error: Invalid user ID in token",
// 		})
// 	}
// 	userID := uint(userIDFloat)

// 	trackingLaporanID := c.Params("id")
// 	if trackingLaporanID == "" {
// 		log.Println("❌ Parameter ID kosong")
// 		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusBadRequest,
// 			Status:  "error",
// 			Message: "ID is required",
// 		})
// 	}

// 	db := database.GetGormDBInstance()
// 	var trackingLaporan models.TrackingLaporan
// 	if err := db.First(&trackingLaporan, trackingLaporanID).Error; err != nil {
// 		log.Printf("❌ Gagal menemukan tracking laporan: %v\n", err)
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return c.Status(http.StatusNotFound).JSON(helper.ResponseWithOutData{
// 				Code:    http.StatusNotFound,
// 				Status:  "error",
// 				Message: "Tracking Laporan not found",
// 			})
// 		}
// 		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Database error",
// 		})
// 	}

// 	if trackingLaporan.Document != nil {
// 		cld, err := cloudinary.NewFromParams("dgnexszl2", "228927731812515", "<your_api_secret>")
// 		if err != nil {
// 			log.Fatalf("Failed to initialize Cloudinary: %v", err)
// 		}
// 		var ctx = context.Background()
// 		var publicIds []string
// 		// Ambil URL gambar dari dokumen
// 		imageURLs := trackingLaporan.Document["urls"].([]interface{})
// 		for _, imageUrl := range imageURLs {
// 			arr := strings.Split(imageUrl.(string), "/")
// 			publicID := arr[len(arr)-1] // Ambil public ID dari URL
// 			publicIds = append(publicIds, publicID)
// 		}
// 		// Hapus gambar dari Cloudinary
// 		_, err = cld.Admin.DeleteAssets(ctx, admin.DeleteAssetsParams{
// 			PublicIDs:    publicIds,
// 			DeliveryType: "upload",
// 			AssetType:    "image",
// 		})
// 		if err != nil {
// 			log.Printf("Failed to delete image: %v", err)
// 			return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
// 				Message: "Failed to delete image",
// 			})
// 		}
// 	}

// 	// Parse body request
// 	var request struct {
// 		NoRegistrasi string   `json:"no_registrasi"`
// 		Keterangan   string   `json:"keterangan"`
// 		Document     []string `json:"document"`
// 	}
// 	if err := c.BodyParser(&request); err != nil {
// 		log.Printf("❌ Gagal parsing body request: %v\n", err)
// 		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusBadRequest,
// 			Status:  "error",
// 			Message: "Invalid request body",
// 		})
// 	}
// 	log.Println("✅ Body request berhasil di-parse")

// 	// Update fields jika ada
// 	if request.NoRegistrasi != "" {
// 		// Validasi no_registrasi
// 		var laporan models.Laporan
// 		if err := db.Where("no_registrasi = ?", request.NoRegistrasi).First(&laporan).Error; err != nil {
// 			log.Printf("❌ No Registrasi %s not found: %v\n", request.NoRegistrasi, err)
// 			return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
// 				Code:    http.StatusBadRequest,
// 				Status:  "error",
// 				Message: "No Registrasi not found in Laporan table",
// 			})
// 		}
// 		trackingLaporan.NoRegistrasi = request.NoRegistrasi
// 	}
// 	if request.Keterangan != "" {
// 		trackingLaporan.Keterangan = request.Keterangan
// 	}
// 	if len(request.Document) > 0 {
// 		trackingLaporan.Document = datatypes.JSONMap{"urls": request.Document}
// 	}

// 	now := time.Now()
// 	trackingLaporan.UpdatedAt = now

// 	// Simpan perubahan
// 	if err := db.Save(&trackingLaporan).Error; err != nil {
// 		log.Printf("❌ Gagal menyimpan perubahan ke database: %v\n", err)
// 		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
// 			Code:    http.StatusInternalServerError,
// 			Status:  "error",
// 			Message: "Failed to update tracking laporan",
// 		})
// 	}

// 	// Notifikasi untuk user
// 	var laporan models.Laporan
// 	if err := db.Where("no_registrasi = ?", trackingLaporan.NoRegistrasi).First(&laporan).Error; err == nil {
// 		var user models.User
// 		if err := db.Where("id = ?", laporan.UserID).First(&user).Error; err == nil && user.NotificationToken != "" {
// 			notificationData := models.FCMNotificationData{
// 				Type:      "tracking_update",
// 				ReportID:  trackingLaporan.NoRegistrasi,
// 				Status:    "updated_tracking",
// 				UpdatedAt: now.Format(time.RFC3339),
// 				Notes:     trackingLaporan.Keterangan,
// 				UpdatedBy: userID,
// 				DeepLink:  "laporanku://tracking/" + trackingLaporan.NoRegistrasi,
// 			}

// 			docMessage := "Ada perubahan terbaru pada tracking laporanmu!"
// 			if len(request.Document) > 0 {
// 				docMessage = "Dokumen baru (PDF/Image) telah diperbarui untuk laporanmu!"
// 			}

// 			notification, err := NewNotificationFromFCMData(
// 				laporan.UserID,
// 				"Tracking Laporanmu Diperbarui!",
// 				"Yay! Tracking untuk laporan No. "+trackingLaporan.NoRegistrasi+" telah diperbarui. "+docMessage+" Yuk cek detailnya!",
// 				notificationData,
// 				now,
// 			)
// 			if err != nil {
// 				log.Printf("Error creating notification: %v", err)
// 			} else {
// 				if err := db.Create(&notification).Error; err != nil {
// 					log.Printf("Failed to store notification: %v", err)
// 				}
// 				if err := SendFCMNotification(user.NotificationToken, notificationData, *notification); err != nil {
// 					log.Printf("Failed to send FCM notification: %v", err)
// 				}
// 			}
// 		}
// 	}

//		response := helper.ResponseWithData{
//			Code:    http.StatusOK,
//			Status:  "success",
//			Message: "Tracking laporan updated successfully",
//			Data: fiber.Map{
//				"id":            trackingLaporan.ID,
//				"no_registrasi": trackingLaporan.NoRegistrasi,
//				"keterangan":    trackingLaporan.Keterangan,
//				"document": fiber.Map{
//					"urls": request.Document,
//				},
//				"created_at": trackingLaporan.CreatedAt,
//				"updated_at": trackingLaporan.UpdatedAt,
//			},
//		}
//		log.Printf("✅ Tracking laporan %s updated successfully by user %d", trackingLaporan.NoRegistrasi, userID)
//		return c.Status(http.StatusOK).JSON(response)
//	}
func DeleteTrackingLaporan(c *fiber.Ctx) error {
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

	trackingLaporanID := c.Params("id")
	if trackingLaporanID == "" {
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "ID is required",
		})
	}

	db := database.GetGormDBInstance()
	var trackingLaporan models.TrackingLaporan
	if err := db.First(&trackingLaporan, trackingLaporanID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(http.StatusNotFound).JSON(helper.ResponseWithOutData{
				Code:    http.StatusNotFound,
				Status:  "error",
				Message: "Tracking Laporan not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Database error",
		})
	}

	noRegistrasi := trackingLaporan.NoRegistrasi
	now := time.Now()

	// Notifikasi sebelum delete
	var laporan models.Laporan
	if err := db.Where("no_registrasi = ?", noRegistrasi).First(&laporan).Error; err == nil {
		var user models.User
		if err := db.Where("id = ?", laporan.UserID).First(&user).Error; err == nil && user.NotificationToken != "" {
			notificationData := models.FCMNotificationData{
				Type:      "tracking_update",
				ReportID:  noRegistrasi,
				Status:    "deleted_tracking",
				UpdatedAt: now.Format(time.RFC3339),
				Notes:     "Tracking Laporan telah dihapus",
				UpdatedBy: userID,
			}

			notification, err := NewNotificationFromFCMData(
				laporan.UserID,
				"Tracking Laporan telah dihapus!",
				"Halo! Tracking untuk laporan No. "+noRegistrasi+" telah dihapus dari sistem. Ada pertanyaan? Hubungi kami ya!",
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
	}

	if err := db.Delete(&trackingLaporan).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to delete tracking laporan",
		})
	}

	return c.Status(http.StatusOK).JSON(helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Tracking laporan deleted successfully",
	})
}
