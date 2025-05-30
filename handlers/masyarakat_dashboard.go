package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func HelloMasyarakat(c *fiber.Ctx) error {
	response := map[string]string{
		"message": "Hai, masyarakat!",
	}
	return c.JSON(response)
}

func EmergencyContact(c *fiber.Ctx) error {
	var emergencyContact models.EmergencyContact
	if err := database.DB.First(&emergencyContact).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to get emergency contact", Data: nil})
	}
	return c.Status(http.StatusOK).JSON(Response{Success: 1, Message: "Emergency contact retrieved successfully", Data: emergencyContact})
}
