package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"strconv"
	"time"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/admin"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportClientRequest struct {
	ClientID       uint64 `json:"client_id" validate:"required"`
	ChatMessageID  string `json:"chat_message_id" validate:"required"`
	MessageContent string `json:"message_content" validate:"required"`
	Notes          string `json:"notes"`
}
type UserReportAdminRequest struct {
	AdminID        uint64 `json:"admin_id" validate:"required"`
	ChatMessageID  string `json:"chat_message_id" validate:"required"`
	MessageContent string `json:"message_content" validate:"required"`
	Notes          string `json:"notes"`
}
type SendPushNotificationRequest struct {
	ClientID uint64                     `json:"client_id" validate:"required"`
	Title    string                     `json:"title" validate:"required"`
	Body     string                     `json:"body" validate:"required"`
	Type     string                     `json:"type" validate:"required"`
	Data     models.FCMNotificationData `json:"data"`
}
type Pagination struct {
	Page     int `query:"page" validate:"min=1"`
	PageSize int `query:"page_size" validate:"min=1,max=100"`
}

func ReportClient(c *fiber.Ctx) error {
	db := database.GetGormDBInstance()
	now := time.Now()

	// JWT Authentication
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{

			Message: "Internal server error: Unable to retrieve user token",
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid token claims",
		})
	}
	fmt.Println("INI NILAI DARI CLAIMS", claims)
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid user ID in token",
		})
	}
	userID := int64(userIDFloat)

	// Validate request body
	var req ReportClientRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Message: "Invalid request body",
		})
	}
	var existingReport models.ReportAdmin
	err := db.Where("reporter_id = ? AND chat_message_id = ? AND report_type = ? AND deleted_at IS NULL",
		userID, req.ChatMessageID, "admin_notification").
		First(&existingReport).Error

	if err == nil {
		// Report sudah ada
		return c.Status(fiber.StatusConflict).JSON(helper.ResponseWithOutData{
			Code:    fiber.StatusConflict,
			Message: "Laporan sudah dibuat untuk pesan ini oleh Anda",
		})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Kesalahan query database
		log.Printf("Database error checking existing report: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    fiber.StatusInternalServerError,
			Message: "Gagal memproses laporan",
		})
	}

	// Generate unique report ID
	reportID := uuid.New().String()

	// Create report record
	report := models.ReportAdmin{
		ReportID:       reportID,
		ReporterID:     userID,
		ReportedUserID: req.ClientID,
		ChatMessageID:  req.ChatMessageID,
		MessageContent: req.MessageContent,
		ReportType:     "admin_notification",
		Status:         "resolved", // Notifications are auto-resolved
		Notes:          req.Notes,
		UpdatedBy:      &userID,
	}

	if err := db.Create(&report).Error; err != nil {
		log.Printf("Failed to store report: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to store report",
		})
	}

	// Fetch client for notification token
	var client models.User
	if err := db.First(&client, req.ClientID).Error; err != nil {
		log.Printf("Client not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(helper.ResponseWithOutData{
			Code:    fiber.StatusNotFound,
			Message: "Client not found",
		})
	}

	// Create and send notification
	notificationData := models.FCMNotificationData{
		Type:      "admin_notification",
		ReportID:  reportID,
		Status:    "resolved",
		UpdatedBy: uint(userID),
		UpdatedAt: now.Format(time.RFC3339),
		Notes:     "Anda telah dilaporkan karena menggunakan kata-kata tidak pantas.",
		DeepLink:  "laporanku://notifications/" + reportID,
	}

	notification, err := NewNotificationFromFCMData(
		uint(req.ClientID),
		"Peringatan: Pelanggaran Etika Komunikasi",
		"Anda telah dilaporkan karena menggunakan kata-kata tidak pantas dalam chat.",
		notificationData,
		now,
	)
	if err != nil {
		log.Printf("Error creating notification: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Failed to create notification",
		})
	}

	// Save notification to database
	if err := db.Create(&notification).Error; err != nil {
		log.Printf("Failed to store notification: %v", err)
	}

	// Send push notification via FCM
	if err := SendFCMNotification(client.NotificationToken, notificationData, *notification); err != nil {
		log.Printf("Failed to send FCM notification: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(helper.ResponseWithData{
		Code:    fiber.StatusOK,
		Message: "Notification sent to client",
		Data:    report,
	})
}

func UserReportAdmin(c *fiber.Ctx) error {
	db := database.GetGormDBInstance()
	now := time.Now()

	// JWT Authentication
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Unable to retrieve user token",
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid token claims",
		})
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid user ID in token",
		})
	}
	userID := int64(userIDFloat)

	// Validate request body
	var req UserReportAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Message: "Invalid request body",
		})
	}

	// Check for duplicate report
	var existingReport models.ReportAdmin
	if err := db.Where("reporter_id = ? AND chat_message_id = ? AND report_type = ?",
		userID, req.ChatMessageID, "client_report").First(&existingReport).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(helper.ResponseWithOutData{
			Message: "This message has already been reported",
		})
	}

	// Generate unique report ID
	reportID := uuid.New().String()
	// Create report record
	report := models.ReportAdmin{
		ReportID:       reportID,
		ReporterID:     int64(userID),
		ReportedUserID: req.AdminID,
		ChatMessageID:  req.ChatMessageID,
		MessageContent: req.MessageContent,
		ReportType:     "client_report",
		Status:         "pending",
		Notes:          req.Notes,
		UpdatedBy:      &userID,
	}

	if err := db.Create(&report).Error; err != nil {
		log.Printf("Failed to store report: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Failed to store report",
		})
	}

	// Fetch client for notification token
	var client models.User
	if err := db.First(&client, userID).Error; err != nil {
		log.Printf("Client not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(helper.ResponseWithOutData{
			Message: "Client not found",
		})
	}

	// Create and send notification
	notificationData := models.FCMNotificationData{
		Type:      "report_status",
		ReportID:  reportID,
		Status:    "pending",
		UpdatedBy: uint(userID),
		UpdatedAt: now.Format(time.RFC3339),
		Notes:     "Terima kasih telah melaporkan admin yang menggunakan kata-kata tidak pantas. Laporan Anda dengan ID " + reportID + " telah diterima dan sedang diverifikasi oleh tim kami.",
		DeepLink:  "laporanku://reports/" + reportID,
	}
	notification, err := NewNotificationFromFCMData(
		uint(userID),
		"Terima Kasih atas Laporan Anda",
		"Terima kasih telah melaporkan admin yang menggunakan kata-kata tidak pantas. Laporan Anda dengan ID "+reportID+" telah diterima dan sedang diverifikasi oleh tim kami.",
		notificationData,
		now,
	)

	if err != nil {
		log.Printf("Error creating notification: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Failed to create notification",
		})
	}

	// Save notification to database
	if err := db.Create(&notification).Error; err != nil {
		log.Printf("Failed to store notification: %v", err)
	}

	// Send push notification via FCM
	if err := SendFCMNotification(client.NotificationToken, notificationData, *notification); err != nil {
		log.Printf("Failed to send FCM notification: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(helper.ResponseWithData{
		Message: "Report submitted successfully",
		Data:    report,
	})
}

func SendPushNotification(c *fiber.Ctx) error {
	db := database.GetGormDBInstance()
	now := time.Now()

	// JWT Authentication
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Unable to retrieve user token",
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid token claims",
		})
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)

	// Validate request body
	var req SendPushNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Message: "Invalid request body",
		})
	}

	// Fetch client for notification token
	var client models.User
	if err := db.First(&client, req.ClientID).Error; err != nil {
		log.Printf("Client not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(helper.ResponseWithOutData{
			Message: "Client not found",
		})
	}

	// Update notification data with defaults
	req.Data.UpdatedBy = userID
	req.Data.UpdatedAt = now.Format(time.RFC3339)
	if req.Data.DeepLink == "" {
		req.Data.DeepLink = "laporanku://notifications/general"
	}
	notificationData := models.FCMNotificationData{
		Type:      "chat_message",
		ReportID:  req.Data.ReportID,
		Status:    "chat_message",
		UpdatedBy: userID,
		UpdatedAt: time.Now().Format(time.RFC3339),
		Notes:     req.Data.Notes,
	}
	// Create notification
	notification, err := NewNotificationFromFCMData(
		uint(req.ClientID),
		req.Title,
		req.Body,
		notificationData,
		now,
	)
	if err != nil {
		log.Printf("Error creating notification: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Failed to create notification",
		})
	}

	// Save notification to database
	if err := db.Create(&notification).Error; err != nil {
		log.Printf("Failed to store notification: %v", err)
	}

	// Send push notification via FCM
	if err := SendFCMNotification(client.NotificationToken, req.Data, *notification); err != nil {
		log.Printf("Failed to send FCM notification: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(helper.ResponseWithOutData{
		Message: "Notification sent successfully",
	})
}

func GetReportedByClient(c *fiber.Ctx) error {
	db := database.GetGormDBInstance()

	// JWT Authentication
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Unable to retrieve user token",
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid token claims",
		})
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)
	log.Printf("Fetching all reports for userID: %d", userID)

	// Parse pagination parameters
	var pagination Pagination
	pagination.Page, _ = strconv.Atoi(c.Query("page", "1"))
	pagination.PageSize, _ = strconv.Atoi(c.Query("page_size", "10"))
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 || pagination.PageSize > 100 {
		pagination.PageSize = 10
	}

	// Fetch reports with usernames
	var reports []struct {
		models.ReportAdmin
		ReporterUsername     string `json:"reporter_username"`
		ReportedUserUsername string `json:"reported_user_username"`
	}
	var total int64
	offset := (pagination.Page - 1) * pagination.PageSize

	// Count total reports
	if err := db.Model(&models.ReportAdmin{}).
		Count(&total).Error; err != nil {
		log.Printf("Failed to count reports: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Failed to fetch reports",
		})
	}

	// Fetch reports with joins to get usernames
	if err := db.Model(&models.ReportAdmin{}).
		Select("report_admins.*, COALESCE(reporter_user.username, 'Unknown') AS reporter_username, COALESCE(reported_user.username, 'Unknown') AS reported_user_username").
		Joins("LEFT JOIN users AS reporter_user ON report_admins.reporter_id = reporter_user.id").
		Joins("LEFT JOIN users AS reported_user ON report_admins.reported_user_id = reported_user.id").
		Offset(offset).
		Limit(pagination.PageSize).
		Find(&reports).Error; err != nil {
		log.Printf("Failed to fetch reports: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Failed to fetch reports",
		})
	}

	log.Printf("Fetched %d reports", len(reports))

	return c.Status(fiber.StatusOK).JSON(helper.ResponseWithData{
		Message: "Reports fetched successfully",
		Data: map[string]interface{}{
			"reports": reports,
			"pagination": map[string]int{
				"page":      pagination.Page,
				"page_size": pagination.PageSize,
				"total":     int(total),
			},
		},
	})
}
func GetUsernameByID(c *fiber.Ctx) error {
	db := database.GetGormDBInstance()

	// JWT Authentication
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Unable to retrieve user token",
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid token claims",
		})
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)
	log.Printf("Fetching username for client ID by userID: %d", userID)

	clientIDStr := c.Params("id")
	clientID, err := strconv.ParseUint(clientIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid client ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Message: "Invalid client ID",
		})
	}

	var user models.User
	if err := db.Select("username").
		Where("id = ?", clientID).
		First(&user).Error; err != nil {
		log.Printf("Failed to fetch username for client ID %d: %v", clientID, err)
		return c.Status(fiber.StatusNotFound).JSON(helper.ResponseWithOutData{
			Message: "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ResponseWithData{
		Message: "Username fetched successfully",
		Data: map[string]interface{}{
			"client_id": clientID,
			"username":  user.Username,
		},
	})
}

func GetUsernamesByIDs(c *fiber.Ctx) error {
	db := database.GetGormDBInstance()

	// JWT Authentication
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Unable to retrieve user token",
		})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid token claims",
		})
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Internal server error: Invalid user ID in token",
		})
	}
	userID := uint(userIDFloat)
	log.Printf("Fetching usernames for userID: %d", userID)

	// Parse body request
	var request struct {
		ClientIDs []uint `json:"client_ids"`
	}
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Message: "Invalid request body",
		})
	}

	if len(request.ClientIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Message: "Client IDs list cannot be empty",
		})
	}

	// Query usernames
	var users []struct {
		ID       uint   `json:"client_id"`
		Username string `json:"username"`
	}
	if err := db.Model(&models.User{}).
		Select("id, username").
		Where("id IN ?", request.ClientIDs).
		Find(&users).Error; err != nil {
		log.Printf("Failed to fetch usernames: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Failed to fetch usernames",
		})
	}

	// Buat map untuk respons
	usernameMap := make(map[uint]string)
	for _, user := range users {
		usernameMap[user.ID] = user.Username
	}

	log.Printf("Fetched %d usernames", len(users))

	return c.Status(fiber.StatusOK).JSON(helper.ResponseWithData{
		Message: "Usernames fetched successfully",
		Data:    usernameMap,
	})
}

func DeleteImage(c *fiber.Ctx) error {

	var request struct {
		ImageUrls []string `json:"image_urls"`
	}
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Message: "Invalid request body",
		})
	}

	if len(request.ImageUrls) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Message: "Public ID cannot be empty",
		})
	}

	cld, err := cloudinary.NewFromParams("dgnexszl2", "228927731812515", "y1js3wYIec7qP8E-sBhkNF7n7Xs")
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	var ctx = context.Background()
	var publicIds []string

	for _, imageUrl := range request.ImageUrls {
		arr := strings.Split(imageUrl, "/")
		publicID := arr[len(arr)-1]
		publicIds = append(publicIds, publicID)
	}

	resp, err := cld.Admin.DeleteAssets(ctx, admin.DeleteAssetsParams{
		PublicIDs:    publicIds,
		DeliveryType: "upload",
		AssetType:    "image",
	})
	if err != nil {
		log.Printf("Failed to delete image: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Message: "Failed to delete image",
		})
	}
	fmt.Println("Response:", resp)
	return c.Status(fiber.StatusOK).JSON(helper.ResponseWithOutData{
		Message: "Image deleted successfully",
	})
}
