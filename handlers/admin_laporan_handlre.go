package handlers

import (
	"backend-pedika-fiber/auth"
	"backend-pedika-fiber/database"
	"strconv"

	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

/*=========================== AMBIL SEMUA LAPORAN =======================*/
func GetLatestReports(c *fiber.Ctx) error {
	var reports []models.Laporan
	db := database.GetGormDBInstance()

	if err := db.
		Preload("ViolenceCategory").
		Order("created_at desc").
		Limit(10).
		Find(&reports).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to fetch latest reports",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	var result []map[string]interface{}
	for _, report := range reports {
		result = append(result, map[string]interface{}{
			"no_registrasi":         report.NoRegistrasi,
			"user_id":               report.UserID,
			"violence_category":     report.ViolenceCategory,
			"kategori_kekerasan_id": report.KategoriKekerasanID,
			"tanggal_pelaporan":     report.TanggalPelaporan,
			"tanggal_kejadian":      report.TanggalKejadian,
			"kategori_lokasi_kasus": report.KategoriLokasiKasus,
			"alamat_tkp":            report.AlamatTKP,
			"alamat_detail_tkp":     report.AlamatDetailTKP,
			"kronologis_kasus":      report.KronologisKasus,
			"status":                report.Status,
			"alasan_dibatalkan":     report.AlasanDibatalkan,
			"waktu_dibatalkan":      report.WaktuDibatalkan,
			"waktu_dilihat":         report.WaktuDilihat,
			"userid_melihat":        report.UserIDMelihat,
			"waktu_diproses":        report.WaktuDiproses,
			"dokumentasi":           report.Dokumentasi,
			"created_at":            report.CreatedAt,
			"updated_at":            report.UpdatedAt,
		})
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Latest reports retrieved successfully",
		Data:    result,
	}
	return c.Status(http.StatusOK).JSON(response)
}


func GetLatestReportsPagination(c *fiber.Ctx) error {
    // Mendapatkan parameter page dan limit dari query string
    page, _ := strconv.Atoi(c.Query("page", "1"))
    limit, _ := strconv.Atoi(c.Query("limit", "10"))
    
    // Menghitung offset
    if page < 1 {
        page = 1
    }
    offset := (page - 1) * limit

    var reports []models.Laporan
    db := database.GetGormDBInstance()

    // Menambahkan total count untuk informasi pagination
    var total int64
    if err := db.Model(&models.Laporan{}).Count(&total).Error; err != nil {
        response := helper.ResponseWithOutData{
            Code:    http.StatusInternalServerError,
            Status:  "error",
            Message: "Failed to fetch total count",
        }
        return c.Status(http.StatusInternalServerError).JSON(response)
    }

    if err := db.
        Preload("ViolenceCategory").
        Order("created_at desc").
        Limit(limit).
        Offset(offset).
        Find(&reports).Error; err != nil {
        response := helper.ResponseWithOutData{
            Code:    http.StatusInternalServerError,
            Status:  "error",
            Message: "Failed to fetch latest reports",
        }
        return c.Status(http.StatusInternalServerError).JSON(response)
    }

    var result []map[string]interface{}
    for _, report := range reports {
        result = append(result, map[string]interface{}{
            "no_registrasi":         report.NoRegistrasi,
            "user_id":               report.UserID,
            "violence_category":     report.ViolenceCategory,
            "kategori_kekerasan_id": report.KategoriKekerasanID,
            "tanggal_pelaporan":     report.TanggalPelaporan,
            "tanggal_kejadian":      report.TanggalKejadian,
            "kategori_lokasi_kasus": report.KategoriLokasiKasus,
            "alamat_tkp":            report.AlamatTKP,
            "alamat_detail_tkp":     report.AlamatDetailTKP,
            "kronologis_kasus":      report.KronologisKasus,
            "status":                report.Status,
            "alasan_dibatalkan":     report.AlasanDibatalkan,
            "waktu_dibatalkan":      report.WaktuDibatalkan,
            "waktu_dilihat":         report.WaktuDilihat,
            "userid_melihat":        report.UserIDMelihat,
            "waktu_diproses":        report.WaktuDiproses,
            "dokumentasi":           report.Dokumentasi,
            "created_at":            report.CreatedAt,
            "updated_at":            report.UpdatedAt,
        })
    }

    // Membuat response dengan metadata pagination
    response := map[string]interface{}{
        "code":    http.StatusOK,
        "status":  "success",
        "message": "Latest reports retrieved successfully",
        "data":    result,
        "meta": map[string]int64{
            "total":    total,
            "page":     int64(page),
            "limit":    int64(limit),
            "totalPage": (total + int64(limit) - 1) / int64(limit),
        },
    }

    return c.Status(http.StatusOK).JSON(response)
}


/*=========================== TAMPILKAN DETAIL LAPORAN USER BERDASARKAN NO_REGISTRASI =======================*/

func GetLaporanByNoRegistrasi(c *fiber.Ctx) error {
	noRegistrasi := c.Params("no_registrasi")
	var laporan models.Laporan
	db := database.GetGormDBInstance()

	if err := db.
		Preload("User").
		Preload("ViolenceCategory").
		Where("no_registrasi = ?", noRegistrasi).
		First(&laporan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response := helper.ResponseWithOutData{
				Code:    http.StatusNotFound,
				Status:  "error",
				Message: "Report not found",
			}
			return c.Status(http.StatusNotFound).JSON(response)
		}
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to fetch report detail",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

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

func AdminLihatLaporan(c *fiber.Ctx) error {
	noRegistrasi := c.Params("no_registrasi")
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

	laporan.Status = "Dilihat"
	now := time.Now()
	laporan.WaktuDilihat = &now
	laporan.UserIDMelihat = &userID

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
		Message: "Laporan status updated successfully",
		Data: fiber.Map{
			"no_registrasi":  laporan.NoRegistrasi,
			"status":         laporan.Status,
			"waktu_dilihat":  laporan.WaktuDilihat,
			"userid_melihat": laporan.UserIDMelihat,
			"updated_at":     laporan.UpdatedAt,
		},
	}

	return c.Status(http.StatusOK).JSON(response)
}

// Custom For push notification when status report is "Diproses"
func AdminProsesLaporan(c *fiber.Ctx) error {
	noRegistrasi := c.Params("no_registrasi")
	
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

	laporan.Status = "Diproses"
	now := time.Now()
	laporan.WaktuDiproses = &now

	if err := db.Save(&laporan).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to update laporan",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

 	var user models.User
    if err := db.Where("id = ?", laporan.UserID).First(&user).Error; err != nil {
        log.Printf("Failed to retrieve user for notification: %v", err)
    }  else if user.NotificationToken != "" {
		log.Println("User notification token:", user.NotificationToken)

		notificationData := models.FCMNotificationData{
			Type:      "report_status",
			ReportID:  laporan.NoRegistrasi,
			Status:    "in_progress",
			UpdatedBy: userID,
			UpdatedAt: now.Format(time.RFC3339),
			Notes:     "Laporan sedang diverifikasi oleh tim",
			// Deep Link Dummy di front End belum digunakan 
			DeepLink:  "laporanku://reports/" + laporan.NoRegistrasi,   
			// ImageURL: ,  INI OPTIONAL tpi tiati harus make size image yang sesuai 
		}

		notification, err := NewNotificationFromFCMData(
			laporan.UserID,
			"Status Laporan Diperbarui",
			"Laporan Anda dengan ID "+laporan.NoRegistrasi+" sedang diproses",
			notificationData,
			now,
		)
		if err != nil {
			log.Println("Error create notif:", err)
			return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Failed to create notification",
			})
		}

		// Simpan notifikasi ke database
		if err := db.Create(&notification).Error; err != nil {
			log.Printf("Failed to store notification: %v", err)
		}

		// Kirim push notification via FCM
		if err := SendFCMNotification(user.NotificationToken, notificationData,*notification); err != nil {
			log.Printf("Failed to send FCM notification: %v", err)
		}
	}

	// Sukses
	return c.Status(http.StatusOK).JSON(helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Laporan Berhasil Diproses",
		Data: fiber.Map{
			"no_registrasi":  laporan.NoRegistrasi,
			"status":         laporan.Status,
			"waktu_diproses": laporan.WaktuDiproses,
			"updated_at":     laporan.UpdatedAt,
		},
	})
}



type StatusStatsResponse struct {
    LaporanMasuk   int `json:"laporanMasuk"`
    LaporanDilihat int `json:"laporanDilihat"`
    LaporanDiproses int `json:"laporanDiproses"`
    LaporanSelesai int `json:"laporanSelesai"`
    LaporanDibatalkan int `json:"laporanDibatalkan"`
}

// Endpoint untuk mendapatkan statistik status laporan
func GetLaporanStatusCount(c *fiber.Ctx) error {

  
    // Query ke database
    db := database.GetGormDBInstance()
    rows, err := db.Raw("SELECT status, COUNT(*) as count FROM laporans GROUP BY status").Rows()
    if err != nil {
        log.Printf("Failed to query status stats: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(helper.ResponseWithOutData{
            Code:    fiber.StatusInternalServerError,
            Status:  "error",
            Message: "Failed to retrieve status statistics",
        })
    }
    defer rows.Close()

  
    stats := StatusStatsResponse{
        LaporanMasuk:      0,
        LaporanDilihat:    0,
        LaporanDiproses:   0,
        LaporanSelesai:    0,
        LaporanDibatalkan: 0,
    }

    
    for rows.Next() {
        var status string
        var count int
        if err := rows.Scan(&status, &count); err != nil {
            log.Printf("Failed to scan row: %v", err)
            continue
        }
        switch status {
        case "Laporan masuk":
            stats.LaporanMasuk = count
        case "Dilihat":
            stats.LaporanDilihat = count
        case "Diproses":
            stats.LaporanDiproses = count
        case "Selesai":
            stats.LaporanSelesai = count
        case "Dibatalkan":
            stats.LaporanDibatalkan = count
        }
    }

    
    return c.Status(fiber.StatusOK).JSON(helper.ResponseWithData{
        Code:    fiber.StatusOK,
        Status:  "success",
        Message: "Status statistics retrieved successfully",
        Data:    stats,
    })
}


func SelesaikanLaporan(c *fiber.Ctx) error {
    noRegistrasi := c.Params("no_registrasi")

    // Ambil token pengguna dari context (admin yang menyelesaikan laporan)
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
    adminID := uint(userIDFloat) // ID admin yang menyelesaikan laporan

    // Cari laporan
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

    // Update status laporan
    laporan.Status = "Selesai"
    now := time.Now()
    laporan.UpdatedAt = now
    if err := db.Save(&laporan).Error; err != nil {
        response := helper.ResponseWithOutData{
            Code:    http.StatusInternalServerError,
            Status:  "error",
            Message: "Failed to update laporan",
        }
        return c.Status(http.StatusInternalServerError).JSON(response)
    }

    // Cari pengguna untuk notifikasi
    var user models.User
    if err := db.Where("id = ?", laporan.UserID).First(&user).Error; err != nil {
        log.Printf("Failed to retrieve user for notification: %v", err)
    } else if user.NotificationToken != "" {
        log.Println("User notification token:", user.NotificationToken)

        // Siapkan data notifikasi FCM
        notificationData := models.FCMNotificationData{
            Type:      "report_status",
            ReportID:  laporan.NoRegistrasi,
            Status:    "completed", // Status berbeda dari "in_progress"
            UpdatedBy: adminID,     // ID admin yang menyelesaikan
            UpdatedAt: now.Format(time.RFC3339),
            Notes:     "Laporan Anda telah selesai diproses",
            DeepLink:  "laporanku://reports/" + laporan.NoRegistrasi, // Deep link opsional
        }

        // Buat notifikasi untuk disimpan di database
        notification, err := NewNotificationFromFCMData(
            laporan.UserID,
            "Laporan Selesai",
            "Laporan Anda dengan ID "+laporan.NoRegistrasi+" telah selesai",
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
            if err := SendFCMNotification(user.NotificationToken, notificationData,*notification); err != nil {
                log.Printf("Failed to send FCM notification: %v", err)
            }
        }
    }

    // Response sukses
    response := helper.ResponseWithData{
        Code:    http.StatusOK,
        Status:  "success",
        Message: "Laporan completed successfully",
        Data: fiber.Map{
            "no_registrasi": laporan.NoRegistrasi,
            "status":        laporan.Status,
            "updated_at":    laporan.UpdatedAt,
        },
    }

    return c.Status(http.StatusOK).JSON(response)
}