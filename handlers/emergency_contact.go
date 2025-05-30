package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func GetEmergencyContact(c *fiber.Ctx) error {
	var emergencyContact models.EmergencyContact
	if err := database.DB.First(&emergencyContact).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to get emergency contact", Data: nil})
	}
	return c.Status(http.StatusOK).JSON(Response{Success: 1, Message: "Emergency contact retrieved successfully", Data: emergencyContact})
}

func UpdateEmergencyContact(c *fiber.Ctx) error {
	var updatedEmergencyContact models.EmergencyContact
	if err := c.BodyParser(&updatedEmergencyContact); err != nil {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: "Invalid request body", Data: nil})
	}

	var existingEmergencyContact models.EmergencyContact
	if err := database.DB.First(&existingEmergencyContact).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(Response{Success: 0, Message: "Kontak Darurat Tidak ditemukan", Data: nil})
	}

	existingEmergencyContact.Phone = updatedEmergencyContact.Phone
	if err := database.DB.Save(&existingEmergencyContact).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Gagal Mengupdate Kontak Darurat", Data: nil})
	}

	return c.Status(http.StatusOK).JSON(Response{Success: 1, Message: "Berhasil Mengupdate Kontak Darurat", Data: existingEmergencyContact})
}

func ShowEmergencyContactByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var emergencyContact models.EmergencyContact
	if err := database.DB.First(&emergencyContact, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(Response{Success: 0, Message: "Emergency contact not found", Data: nil})
	}

	return c.Status(http.StatusOK).JSON(Response{Success: 1, Message: "Emergency contact retrieved successfully", Data: emergencyContact})
}
