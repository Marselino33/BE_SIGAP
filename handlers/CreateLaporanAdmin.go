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
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CreateLaporanAdmin allows admin to create a report with file uploads to Cloudinary
func CreateLaporanAdmin(c *fiber.Ctx) error {
	fmt.Println("REQUEST MASUK KE ADMIN")

	token := c.Get("Authorization")
	adminID, err := auth.ExtractUserIDFromToken(token)
	if err != nil {
		fmt.Println("Unauthorized access attempt")
		return c.Status(http.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized",
		})
	}
	fmt.Println("Admin ID extracted:", adminID)

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println("Failed to parse multipart form:", err)
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Failed to parse multipart form",
		})
	}
	fmt.Println("Multipart form parsed successfully")

	// Extract fields
	kategoriVals := form.Value["kategori_kekerasan_id"]
	tglVals := form.Value["tanggal_kejadian"]
	lokasiVals := form.Value["kategori_lokasi_kasus"]
	alamatVals := form.Value["alamat_tkp"]
	detailVals := form.Value["alamat_detail_tkp"]
	kronoVals := form.Value["kronologis_kasus"]

	// Validate required
	if len(kategoriVals) == 0 || len(tglVals) == 0 || len(lokasiVals) == 0 || len(alamatVals) == 0 || len(kronoVals) == 0 {
		fmt.Println("Missing required fields")
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Missing required fields",
		})
	}
	fmt.Println("All required fields are present")

	// Validate and fetch category
	catID, err := strconv.ParseUint(kategoriVals[0], 10, 64)
	if err != nil {
		fmt.Println("Invalid kategori_kekerasan_id:", kategoriVals[0])
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid kategori_kekerasan_id",
		})
	}
	fmt.Println("Parsed kategori_kekerasan_id:", catID)

	var cat models.ViolenceCategory
	if err := database.GetGormDBInstance().First(&cat, catID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("Kategori kekerasan tidak ditemukan:", catID)
			return c.Status(http.StatusNotFound).JSON(helper.ResponseWithOutData{
				Code:    http.StatusNotFound,
				Status:  "error",
				Message: "Kategori kekerasan tidak ditemukan",
			})
		}
		fmt.Println("Database error:", err)
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Database error",
		})
	}
	fmt.Println("Fetched category:", cat)

	// Parse tanggal kejadian
	tgl, err := time.Parse("2006-01-02", tglVals[0])
	if err != nil {
		fmt.Println("Invalid date format for tanggal_kejadian:", tglVals[0])
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Format tanggal kejadian harus YYYY-MM-DD",
		})
	}
	fmt.Println("Parsed tanggal kejadian:", tgl)

	// Upload dokumentasi to Cloudinary
	files := form.File["dokumentasi"]
	var imageURLs []string
	if len(files) > 0 {
		upload, upErr := helper.UploadMultipleFileToCloudinary(files)
		if upErr != nil {
			fmt.Println("Failed to upload dokumentasi:", upErr)
			return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal upload dokumentasi",
			})
		}
		imageURLs = upload
	}
	fmt.Println("Result image URLs:", imageURLs)

	// Generate nomor registrasi
	year := time.Now().Year()
	month := int(time.Now().Month())
	noReg, genErr := generateUniqueNoRegistrasi(month, year)
	if genErr != nil {
		fmt.Println("Failed to generate nomor registrasi:", genErr)
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal generate nomor registrasi",
		})
	}
	fmt.Println("Generated nomor registrasi:", noReg)

	// Build laporan model
	laporan := models.Laporan{
		NoRegistrasi:        noReg,
		UserID:              uint(adminID),
		KategoriKekerasanID: uint(catID),
		TanggalPelaporan:    time.Now(),
		TanggalKejadian:     tgl,
		KategoriLokasiKasus: lokasiVals[0],
		AlamatTKP:           alamatVals[0],
		AlamatDetailTKP:     "",
		KronologisKasus:     kronoVals[0],
		Status:              "Laporan masuk",
		Dokumentasi:         datatypes.JSONMap{"urls": imageURLs},
	}
	if len(detailVals) > 0 {
		laporan.AlamatDetailTKP = detailVals[0]
	}
	fmt.Println("Laporan model built:", laporan)

	// Save to DB
	if err := database.GetGormDBInstance().Create(&laporan).Error; err != nil {
		fmt.Println("Failed to save laporan to database:", err)
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal menyimpan laporan",
		})
	}
	fmt.Println("Laporan successfully saved:", laporan)

	return c.Status(http.StatusCreated).JSON(helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Laporan berhasil dibuat",
		Data:    laporan,
	})
}

func generateUniqueNoRegistrasimin(month, year int) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	ram := convertToRoman(month)
	base := fmt.Sprintf("001-DPMDPPA-%s-%d", ram, year)
	db := database.GetGormDBInstance()
	var cnt int64
	if err := db.Model(&models.Laporan{}).Where("no_registrasi = ?", base).Count(&cnt).Error; err != nil {
		return "", err
	}
	if cnt == 0 {
		return base, nil
	}
	for i := 2; i < 1000; i++ {
		cand := fmt.Sprintf("%03d-DPMDPPA-%s-%d", i, ram, year)
		if err := db.Model(&models.Laporan{}).Where("no_registrasi = ?", cand).Count(&cnt).Error; err != nil {
			return "", err
		}
		if cnt == 0 {
			return cand, nil
		}
	}
	return "", errors.New("unique nomor gagal")
}

func convertToRomans(month int) string {
	rom := [...]string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	if month >= 1 && month <= 12 {
		return rom[month-1]
	}
	return ""
}
