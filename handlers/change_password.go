package handlers

import (
	"backend-pedika-fiber/auth"
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func ChangePassword(c *fiber.Ctx) error {
	// Parse the request body
	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate new password and confirmation password
	if req.NewPassword != req.ConfirmPassword {
		return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "New password and confirmation password do not match",
		})
	}

	// Extract user ID from the token
	userID, err := auth.ExtractUserIDFromToken(c.Get("Authorization"))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to get user ID",
		})
	}

	// Retrieve the user from the database
	user, err := getUserID(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "User not found",
		})
	}

	// Debugging: Print user details and the provided old password
	// REMOVE THESE LINES IN PRODUCTION
	log.Println("Stored password hash:", user.Password)
	log.Println("Provided old password:", req.OldPassword)

	// Compare old password with the stored password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(helper.ResponseWithOutData{
			Code:    http.StatusUnauthorized,
			Status:  "error",
			Message: "Old password is incorrect",
		})
	}

	// Hash the new password
	hashedNewPassword, err := HashPassword(req.NewPassword)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to hash new password",
		})
	}

	// Update the password in the database
	err = updatePasswordInDatabase(userID, hashedNewPassword)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to update password",
		})
	}

	return c.Status(http.StatusOK).JSON(helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Password changed successfully",
	})
}

func getUserID(userID uint) (models.User, error) {
	db := database.GetGormDBInstance()

	var user models.User
	err := db.First(&user, userID).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func updatePasswordInDatabase(userID uint, newPassword string) error {
	db := database.GetGormDBInstance()

	err := db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":   newPassword,
		"updated_at": time.Now(),
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return "", err
	}
	return string(hashedPassword), nil
}
