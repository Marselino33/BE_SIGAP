package handlers

import (
	"backend-pedika-fiber/auth"
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

/*=========================== USER CREATE LAPORAN =======================*/
var mu sync.Mutex

func CreateLaporan(c *fiber.Ctx) error {
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

	var laporan models.Laporan
	if err := c.BodyParser(&laporan); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	categoryViolenceID, err := strconv.ParseUint(c.FormValue("kategori_kekerasan_id"), 10, 64)
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid KategoriKekerasan ID",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	var violenceCategory models.ViolenceCategory
	if err := database.GetGormDBInstance().First(&violenceCategory, categoryViolenceID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Kategori kekerasan yang anda pilih tidak ditemukan",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve multipart form",
		})
	}
	imageURLs := form.Value["dokumentasi"]
	// imageURLs, err := helper.UploadMultipleFileToCloudinary(files)
	// if err != nil {
	// 	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": "Failed to upload images",
	// 	})
	// }

	laporan.Dokumentasi = datatypes.JSONMap{"urls": imageURLs}

	tanggalKejadian, err := time.Parse("2006-01-02T15:04:05", c.FormValue("tanggal_kejadian"))
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid format for tanggal kejadian",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	year := time.Now().Year()
	month := int(time.Now().Month())
	noRegistrasi, err := generateUniqueNoRegistrasi(month, year)
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to generate registration number",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	laporan.NoRegistrasi = noRegistrasi
	laporan.TanggalPelaporan = time.Now()
	laporan.TanggalKejadian = tanggalKejadian
	laporan.KategoriLokasiKasus = c.FormValue("kategori_lokasi_kasus")
	laporan.AlamatTKP = c.FormValue("alamat_tkp")
	laporan.AlamatDetailTKP = c.FormValue("alamat_detail_tkp")
	laporan.KronologisKasus = c.FormValue("kronologis_kasus")
	laporan.Status = "Laporan masuk"
	laporan.KategoriKekerasanID = uint(categoryViolenceID)
	laporan.UserID = uint(userID)
	laporan.CreatedAt = time.Now()
	laporan.UpdatedAt = time.Now()
	laporan.AlasanDibatalkan = ""
	laporan.WaktuDilihat = nil
	laporan.WaktuDiproses = nil
	laporan.WaktuDibatalkan = nil
	laporan.UserIDMelihat = nil

	if err := database.GetGormDBInstance().Create(&laporan).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to create laporan",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Laporan created successfully",
		Data: fiber.Map{
			"no_registrasi":         laporan.NoRegistrasi,
			"user_id":               laporan.UserID,
			"kategori_kekerasan_id": laporan.KategoriKekerasanID,
			"tanggal_pelaporan":     laporan.TanggalPelaporan,
			"tanggal_kejadian":      laporan.TanggalKejadian,
			"kategori_lokasi_kasus": laporan.KategoriLokasiKasus,
			"alamat_tkp":            laporan.AlamatTKP,
			"alamat_detail_tkp":     laporan.AlamatDetailTKP,
			"kronologis_kasus":      laporan.KronologisKasus,
			"dokumentasi": fiber.Map{
				"urls": imageURLs,
			},
			"created_at": laporan.CreatedAt,
			"updated_at": laporan.UpdatedAt,
		},
	}

	return c.Status(http.StatusCreated).JSON(response)
}

func generateUniqueNoRegistrasi(month, year int) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	romanMonth := convertToRoman(month)
	regNo := "001-DPMDPPA-" + romanMonth + "-" + strconv.Itoa(year)
	var existingCount int64
	if err := database.GetGormDBInstance().Model(&models.Laporan{}).Where("no_registrasi = ?", regNo).Count(&existingCount).Error; err != nil {
		return "", err
	}
	if existingCount > 0 {
		for i := 1; i < 1000; i++ {
			modifiedRegNo := fmt.Sprintf("%03d", i) + "-DPMDPPA-" + romanMonth + "-" + strconv.Itoa(year)
			var existingCount int64
			if err := database.GetGormDBInstance().Model(&models.Laporan{}).Where("no_registrasi = ?", modifiedRegNo).Count(&existingCount).Error; err != nil {
				return "", err
			}
			if existingCount == 0 {
				return modifiedRegNo, nil
			}
		}
		return "", errors.New("failed to generate unique registration number")
	}
	return regNo, nil
}

func convertToRoman(month int) string {
	months := [...]string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	if month >= 1 && month <= 12 {
		return months[month-1]
	}
	return ""
}

/*=========================== USER EDIT LAPORAN =======================*/

func EditLaporan(c *fiber.Ctx) error {
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

	noRegistrasi := c.Params("no_registrasi")

	var laporan models.Laporan
	if err := database.GetGormDBInstance().Where("no_registrasi = ?", noRegistrasi).First(&laporan).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Laporan not found",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}

	if laporan.UserID != uint(userID) {
		response := helper.ResponseWithOutData{
			Code:    http.StatusForbidden,
			Status:  "error",
			Message: "You are not authorized to edit this laporan",
		}
		return c.Status(http.StatusForbidden).JSON(response)
	}

	if err := c.BodyParser(&laporan); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	if newCategoryID := c.FormValue("kategori_kekerasan_id"); newCategoryID != "" {
		categoryViolenceID, err := strconv.ParseUint(newCategoryID, 10, 64)
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusBadRequest,
				Status:  "error",
				Message: "Invalid KategoriKekerasan ID",
			}
			return c.Status(http.StatusBadRequest).JSON(response)
		}

		var violenceCategory models.ViolenceCategory
		if err := database.GetGormDBInstance().First(&violenceCategory, categoryViolenceID).Error; err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusNotFound,
				Status:  "error",
				Message: "Violence category not found",
			}
			return c.Status(http.StatusNotFound).JSON(response)
		}
		laporan.KategoriKekerasanID = uint(categoryViolenceID)
	}

	tanggalKejadian := c.FormValue("tanggal_kejadian")
	if tanggalKejadian != "" {
		parsedTanggalKejadian, err := time.Parse("2006-01-02T15:04:05", tanggalKejadian)
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusBadRequest,
				Status:  "error",
				Message: "Invalid format for tanggal kejadian",
			}
			return c.Status(http.StatusBadRequest).JSON(response)
		}
		laporan.TanggalKejadian = parsedTanggalKejadian
	}

	form, err := c.MultipartForm()
	if err == nil && form.File != nil && len(form.File["dokumentasi"]) > 0 {
		files := form.File["dokumentasi"]
		imageURLs, err := helper.UploadMultipleFileToCloudinary(files)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload images",
			})
		}
		laporan.Dokumentasi = datatypes.JSONMap{"urls": imageURLs}
	}

	laporan.KategoriLokasiKasus = c.FormValue("kategori_lokasi_kasus")
	laporan.AlamatTKP = c.FormValue("alamat_tkp")
	laporan.AlamatDetailTKP = c.FormValue("alamat_detail_tkp")
	laporan.KronologisKasus = c.FormValue("kronologis_kasus")
	laporan.UpdatedAt = time.Now()

	if err := database.GetGormDBInstance().Save(&laporan).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to update laporan",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Laporan updated successfully",
		Data:    laporan,
	}

	return c.Status(http.StatusOK).JSON(response)
}

/*=========================== AMBIL SEMUA  LAPORAN SETIAP BERDASARKAN USER YANG LOGIN=======================*/
func GetUserReports(c *fiber.Ctx) error {
	userID, err := auth.ExtractUserIDFromToken(c.Get("Authorization"))
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized",
		}
		return c.Status(http.StatusUnauthorized).JSON(response)
	}
	reports, err := GetReportsByUserID(userID)
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to get user reports",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	db := database.GetGormDBInstance()
	var formattedReports []map[string]interface{}
	for _, report := range reports {
		var violenceCategory models.ViolenceCategory
		if err := db.Where("id = ?", report["kategori_kekerasan_id"]).First(&violenceCategory).Error; err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Failed to fetch violence category detail",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		report["violence_category_detail"] = violenceCategory
		formattedReports = append(formattedReports, report)
	}
	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "List of laporan by user",
		Data:    formattedReports,
	}

	return c.Status(http.StatusOK).JSON(response)
}

func GetReportsByUserID(userID uint) ([]map[string]interface{}, error) {
	var reports []models.Laporan
	if err := database.GetGormDBInstance().
		Where("user_id = ?", userID).
		Find(&reports).Error; err != nil {
		return nil, err
	}

	var formattedReports []map[string]interface{}
	for _, report := range reports {
		formattedReport := map[string]interface{}{
			"no_registrasi":         report.NoRegistrasi,
			"user_id":               report.UserID,
			"kategori_kekerasan_id": report.KategoriKekerasanID,
			"tanggal_pelaporan":     report.TanggalPelaporan,
			"tanggal_kejadian":      report.TanggalKejadian,
			"kategori_lokasi_kasus": report.KategoriLokasiKasus,
			"alamat_tkp":            report.AlamatTKP,
			"alamat_detail_tkp":     report.AlamatDetailTKP,
			"kronologis_kasus":      report.KronologisKasus,
			"status":                report.Status,
			"alasan_dibatalkan":     report.AlasanDibatalkan,
			"waktu_dilihat":         report.WaktuDilihat,
			"userid_melihat":        report.UserIDMelihat,
			"waktu_diproses":        report.WaktuDiproses,
			"waktu_dibatalkan":      report.WaktuDibatalkan,
			"dokumentasi":           report.Dokumentasi,
			"created_at":            report.CreatedAt,
			"updated_at":            report.UpdatedAt,
		}
		formattedReports = append(formattedReports, formattedReport)
	}

	return formattedReports, nil
}

/*=========================== TAMPILKAN DETAIL LAPORAN USER BERDASARKAN NO_REGISTRASI =======================*/
func GetReportByNoRegistrasi(c *fiber.Ctx) error {
	noRegistrasi := c.Params("no_registrasi")
	var laporan models.Laporan
	db := database.GetGormDBInstance()

	// Preload the necessary related data
	if err := db.
		Preload("User").
		Preload("ViolenceCategory").
		Where("no_registrasi = ?", noRegistrasi).
		First(&laporan).Error; err != nil {
		status := http.StatusInternalServerError
		message := "Failed to fetch report detail"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			message = "Report not found"
		}
		response := helper.ResponseWithOutData{
			Code:    status,
			Status:  "error",
			Message: message,
		}
		return c.Status(status).JSON(response)
	}

	// Fetch tracking laporan details
	var trackingLaporan []models.TrackingLaporan
	if err := db.Where("no_registrasi = ?", noRegistrasi).Find(&trackingLaporan).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to fetch tracking laporan details",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	var pelaku []models.Pelaku
	if err := db.Where("no_registrasi = ?", noRegistrasi).Find(&pelaku).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to fetch pelaku details",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	var korban []models.Korban
	if err := db.Where("no_registrasi = ?", noRegistrasi).Find(&korban).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to fetch korban details",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	var userMelihat models.User
	if laporan.UserIDMelihat != nil {
		if err := db.First(&userMelihat, *laporan.UserIDMelihat).Error; err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Failed to fetch user detail who viewed the report",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
	}
	responseData := struct {
		models.Laporan
		TrackingLaporan []models.TrackingLaporan `json:"tracking_laporan"`
		Pelaku          []models.Pelaku          `json:"pelaku"`
		Korban          []models.Korban          `json:"korban"`
		UserMelihat     *models.User             `json:"user_melihat,omitempty"`
	}{
		Laporan:         laporan,
		TrackingLaporan: trackingLaporan,
		Pelaku:          pelaku,
		Korban:          korban,
		UserMelihat:     nil,
	}
	if laporan.UserIDMelihat != nil {
		responseData.UserMelihat = &userMelihat
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Report detail retrieved successfully",
		Data:    responseData,
	}

	return c.Status(http.StatusOK).JSON(response)
}

/*=========================== BATALKAN LAPORAN BERDASARKAN NO_REGISTRASI =======================*/
func BatalkanLaporan(c *fiber.Ctx) error {
	noRegistrasi := c.Params("no_registrasi")

	var laporan models.Laporan
	db := database.GetGormDBInstance()
	if err := db.Where("no_registrasi = ?", noRegistrasi).First(&laporan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response := helper.ResponseWithOutData{
				Code:    http.StatusNotFound,
				Status:  "error",
				Message: "Laporan not found",
			}
			return c.Status(http.StatusNotFound).JSON(response)
		}
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to retrieve laporan",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	alasanDibatalkan := c.FormValue("alasan_dibatalkan")
	if alasanDibatalkan == "" {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Alasan dibatalkan is required",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	laporan.Status = "Dibatalkan"
	laporan.AlasanDibatalkan = alasanDibatalkan
	now := time.Now()
	laporan.WaktuDibatalkan = &now

	if err := db.Save(&laporan).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to update laporan",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Laporan cancelled successfully",
		Data: fiber.Map{
			"no_registrasi":     laporan.NoRegistrasi,
			"status":            laporan.Status,
			"alasan_dibatalkan": laporan.AlasanDibatalkan,
			"waktu_dibatalkan":  laporan.WaktuDibatalkan,
			"updated_at":        laporan.UpdatedAt,
		},
	}

	return c.Status(http.StatusOK).JSON(response)
}
