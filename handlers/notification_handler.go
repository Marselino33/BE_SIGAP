package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

func NewNotificationFromFCMData(userID uint, title, body string, data models.FCMNotificationData, now time.Time) (*models.Notification, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FCMNotificationData: %w", err)
	}

	return &models.Notification{
		UserID:    userID,
		Type:      data.Type,
		Title:     title,
		Body:      body,
		Data:      string(dataJSON),
		IsRead:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// func SendFCMNotification(token string, data models.FCMNotificationData, notification models.Notification) error {
//     ctx := context.Background()

//     opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
//     app, err := firebase.NewApp(ctx, nil, opt)
//     if err != nil {
//         log.Printf("Error initializing Firebase app: %v", err)
//         return err
//     }

//     client, err := app.Messaging(ctx)
//     if err != nil {
//         log.Printf("Error getting Messaging client: %v", err)
//         return err
//     }

//     message := &messaging.Message{
//         Token: token,
//         Data: map[string]string{
//             "type":      data.Type,
//             "reportId":  data.ReportID,
//             "status":    data.Status,
//             "updatedBy": fmt.Sprintf("%d", data.UpdatedBy),
//             "updatedAt": data.UpdatedAt,
//             "notes":     data.Notes,
//             "deepLink":  data.DeepLink,
//             "imageUrl":  data.ImageURL,
//         },
//         Notification: &messaging.Notification{
//             Title:notification.Title,
//             Body:  notification.Body,
//         },
//     }

//     response, err := client.Send(ctx, message)
//     if err != nil {
//         log.Printf("Error sending FCM message: %v", err)
//         return err
//     }

//	    log.Printf("Successfully sent FCM message: %s", response)
//	    return nil
//	}

func getAccessToken(credentialsPath string) (string, error) {
	ctx := context.Background()

	data, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file kredensial: %w", err)
	}

	creds, err := google.CredentialsFromJSON(ctx, data, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return "", fmt.Errorf("gagal mengambil credentials dari JSON: %w", err)
	}

	tokenSource := creds.TokenSource
	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("gagal mengambil token akses: %w", err)
	}

	return token.AccessToken, nil
}
func SendFCMNotification(token string, data models.FCMNotificationData, notification models.Notification) error {
	ctx := context.Background()
	log.Println("Memulai proses pengiriman notifikasi FCM...")

	// Validasi file kredensial
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		log.Printf("File kredensial tidak ditemukan: %s", credentialsPath)
		return fmt.Errorf("file kredensial tidak ditemukan: %s", credentialsPath)
	}
	log.Println("File kredensial ditemukan.")
	accessToken, err := getAccessToken(credentialsPath)
	if err != nil {
		log.Printf("Gagal mendapatkan access token: %v", err)
	} else {
		log.Printf("Access Token (DEBUG): %s", accessToken)
	}
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		log.Printf("File kredensial tidak ditemukan: %s", credentialsPath)
		return fmt.Errorf("file kredensial tidak ditemukan: %s", credentialsPath)
	}
	log.Println("File kredensial ditemukan.")

	// Inisialisasi aplikasi Firebase
	log.Println("Menginisialisasi aplikasi Firebase...")
	opt := option.WithCredentialsFile(credentialsPath)
	config := &firebase.Config{
		ProjectID: "pa2-kelompok07",
	}
	fmt.Println("opt:", opt)
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Printf("Gagal menginisialisasi aplikasi Firebase: %v", err)
		return fmt.Errorf("gagal menginisialisasi aplikasi Firebase: %w", err)
	}
	log.Println("Firebase berhasil diinisialisasi.")

	// Dapatkan klien Messaging
	log.Println("Mendapatkan klien Messaging...")
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("Gagal mendapatkan klien Messaging: %v", err)
		return fmt.Errorf("gagal mendapatkan klien Messaging: %w", err)
	}
	log.Println("Klien Messaging berhasil didapatkan.")

	// Validasi token
	if token == "" {
		log.Println("Token FCM kosong.")
		return fmt.Errorf("token FCM kosong")
	}
	log.Printf("Token FCM valid: %s", token)

	// Log isi data dan notifikasi
	log.Printf("Data notifikasi: %+v", data)
	log.Printf("Payload notifikasi: Title: %s, Body: %s", notification.Title, notification.Body)

	// Buat pesan FCM
	log.Println("Membuat pesan FCM...")
	message := &messaging.Message{
		Token: token,
		Data: map[string]string{
			"type":      data.Type,
			"reportId":  data.ReportID,
			"status":    data.Status,
			"updatedBy": fmt.Sprintf("%d", data.UpdatedBy),
			"updatedAt": data.UpdatedAt,
			"notes":     data.Notes,
			"deepLink":  data.DeepLink,
			"imageUrl":  data.ImageURL,
		},
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
	}
	log.Printf("Pesan FCM berhasil dibuat: %+v", message)

	// Kirim pesan FCM
	log.Println("Mengirim pesan FCM...")
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Printf("Gagal mengirim pesan FCM: %v", err)
		if apiErr, ok := err.(*googleapi.Error); ok {
			log.Printf("Kode Status HTTP: %d", apiErr.Code)
			log.Printf("Detail Kesalahan: %v", apiErr.Details)
		}
		return fmt.Errorf("gagal mengirim pesan FCM: %w", err)
	}

	log.Printf("Berhasil mengirim pesan FCM, response ID: %s", response)
	return nil
}
func StoreNotification(userID uint, notificationData models.FCMNotificationData) error {
	db := database.GetGormDBInstance()

	dataJSON, err := json.Marshal(notificationData)
	if err != nil {
		log.Printf("Failed to marshal notification data: %v", err)
		return err
	}

	notification := models.Notification{
		UserID:    userID,
		Type:      notificationData.Type,
		Title:     "Status Laporan Diperbarui",
		Body:      "Laporan Anda dengan ID " + notificationData.ReportID + " sedang diproses",
		Data:      string(dataJSON),
		IsRead:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&notification).Error; err != nil {
		log.Printf("Failed to store notification: %v", err)
		return err
	}

	return nil
}

func GetUserNotifications(c *fiber.Ctx) error {
	// Ambil token pengguna dari context
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Unable to retrieve user token",
		})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid token claims",
		})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)

	// Ambil parameter query untuk pagination dan filter
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	isRead := c.Query("is_read", "")

	offset := (page - 1) * limit

	// Query database
	db := database.GetGormDBInstance()
	var notifications []models.Notification
	query := db.Where("user_id = ?", userID)

	// Tambahkan filter is_read jika ada
	if isRead != "" {
		readBool, err := strconv.ParseBool(isRead)
		if err == nil {
			query = query.Where("is_read = ?", readBool)
		}
	}

	// Hitung total notifikasi untuk pagination
	var total int64
	if err := query.Model(&models.Notification{}).Count(&total).Error; err != nil {
		log.Printf("Failed to count notifications: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to retrieve notifications count",
		})
	}

	// Ambil data notifikasi dengan limit dan offset
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications).Error; err != nil {
		log.Printf("Failed to retrieve notifications: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to retrieve notifications",
		})
	}

	// Response sukses
	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Notifications retrieved successfully",
		Data: fiber.Map{
			"notifications": notifications,
			"pagination": fiber.Map{
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	}

	return c.Status(http.StatusOK).JSON(response)
}
func GetUserNotificationsAndMarkAsRead(c *fiber.Ctx) error {
	// Ambil token pengguna dari context
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Unable to retrieve user token",
		})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid token claims",
		})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)

	// Ambil parameter query untuk pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	// Query database
	db := database.GetGormDBInstance()
	var notifications []*models.Notification // Ubah ke slice of pointers

	// Hitung total notifikasi untuk pagination (tanpa filter is_read)
	var total int64
	if err := db.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		log.Printf("Failed to count notifications: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to retrieve notifications count",
		})
	}

	// Ambil semua data notifikasi dengan limit dan offset
	if err := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		log.Printf("Failed to retrieve notifications: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to retrieve notifications",
		})
	}

	// Ubah is_read menjadi true untuk notifikasi yang belum dibaca dan update database
	for _, notification := range notifications {
		if !notification.IsRead {
			notification.IsRead = true
			notification.UpdatedAt = time.Now()
			// Update ke database
			if err := db.Model(&models.Notification{}).
				Where("id = ?", notification.ID).
				Updates(map[string]interface{}{
					"is_read":    true,
					"updated_at": notification.UpdatedAt,
				}).Error; err != nil {
				log.Printf("Failed to update is_read for notification %d: %v", notification.ID, err)
			}
		}
	}

	// Response sukses
	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Notifications retrieved successfully",
		Data: fiber.Map{
			"notifications": notifications,
			"pagination": fiber.Map{
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	}

	return c.Status(http.StatusOK).JSON(response)
}

func MarkNotificationAsRead(c *fiber.Ctx) error {
	// Ambil token pengguna dari context
	log.Println("[DEBUG] Content-Type:", c.Get("Content-Type"))
	log.Println("[DEBUG] Raw Query notification_id:", string(c.Query("notification_id")))

	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Unable to retrieve user token",
		})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid token claims",
		})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)

	notificationID := c.Query("notification_id")
	if notificationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Notification ID is required",
		})
	}

	// Query database
	db := database.GetGormDBInstance()
	var notification models.Notification
	if err := db.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Notification not found or not owned by user",
		})
	}

	// Update status is_read
	notification.IsRead = true
	if err := db.Save(&notification).Error; err != nil {
		log.Printf("Failed to mark notification as read: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to mark notification as read",
		})
	}

	// Response sukses
	response := helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Notification marked as read",
	}
	return c.Status(http.StatusOK).JSON(response)
}

func GetUnreadNotificationsCount(c *fiber.Ctx) error {
	// Ambil token pengguna dari context
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Unable to retrieve user token",
		})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid token claims",
		})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)

	// Query database
	db := database.GetGormDBInstance()
	var count int64
	if err := db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		log.Printf("Failed to count unread notifications: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to retrieve unread notifications count",
		})
	}

	// Response sukses
	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Unread notifications count retrieved successfully",
		Data: fiber.Map{
			"unread_count": count,
		},
	}
	return c.Status(http.StatusOK).JSON(response)
}
