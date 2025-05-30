package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type ForgotPasswordRequest struct {
	Email           string `json:"email"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func ForgotPassword(c *fiber.Ctx) error {
	// 1. Parse request
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Body JSON tidak valid",
		})
	}

	// 2. Validasi password baru
	if req.NewPassword != req.ConfirmPassword {
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Password dan konfirmasi tidak cocok",
		})
	}

	// 3. Cari user berdasar email
	db := database.GetGormDBInstance()
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Email tidak terdaftar",
		})
	}

	// 4. Hash password baru
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal memproses password",
		})
	}

	// 5. Update email & password
	//    — jika Anda ingin update email juga, tambahkan field Email di request
	//    dan uncomment baris berikut:
	// user.Email = req.NewEmail

	user.Password = string(hashed)
	user.UpdatedAt = time.Now()

	if err := db.Save(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal mengubah data user",
		})
	}

	// 6. Respon sukses
	return c.Status(http.StatusOK).JSON(helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Password berhasil diubah",
	})
}
