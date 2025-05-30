package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetAllEvent(c *fiber.Ctx) error {
	var event []models.Event
	if err := database.DB.Find(&event).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to retrieve contents",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "List of contents",
		Data:    event,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func GetEventByID(c *fiber.Ctx) error {
	eventID := c.Params("id")

	var event models.Event
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Event not found",
		})
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Event details",
		Data:    event,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func CreateEvent(c *fiber.Ctx) error {
	var event models.Event
	if err := c.BodyParser(&event); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	file, err := c.FormFile("thumbnail_event")
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Image file not provided",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	src, err := file.Open()
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to open image file",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	defer src.Close()

	imageURL, err := helper.UploadFileToCloudinary(src, file.Filename)
	if err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to upload image",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	event.ThumbnailEvent = imageURL
	event.NamaEvent = c.FormValue("nama_event")
	event.DeskripsiEvent = c.FormValue("deskripsi_event")
	tanggalPelaksanaanStr := c.FormValue("tanggal_pelaksanaan")

	var parsedDate time.Time
	var parseErr error

	formats := []string{
		"2006-01-02T15:04",    // Format expected from datetime-local input
		"2006-01-02 15:04:00", // Format received in the log
		"2006-01-02 15:04",    // Without seconds
		"02/01/2006 15:04",    // Another possible format
	}

	for _, format := range formats {
		parsedDate, parseErr = time.Parse(format, tanggalPelaksanaanStr)
		if parseErr == nil {
			break
		}
		fmt.Println("Error parsing date with format", format, ":", parseErr)
	}

	if parseErr != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Format tanggal tidak valid. Format yang diharapkan adalah YYYY-MM-DDTHH:MM",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}

	event.TanggalPelaksanaan = parsedDate

	if err := database.DB.Create(&event).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to create event",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Berhasil membuat Content",
		Data:    event,
	}
	return c.Status(http.StatusCreated).JSON(response)
}

func UpdateEvent(c *fiber.Ctx) error {
	eventID := c.Params("id")

	// Fetch existing event from the database
	var existingEvent models.Event
	if err := database.DB.First(&existingEvent, eventID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Event tidak ditemukan",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}

	// Parse form data for updating the event
	eventName := c.FormValue("nama_event")
	if eventName != "" {
		existingEvent.NamaEvent = eventName
	}

	eventDescription := c.FormValue("deskripsi_event")
	if eventDescription != "" {
		existingEvent.DeskripsiEvent = eventDescription
	}

	eventDate := c.FormValue("tanggal_pelaksanaan")
	if eventDate != "" {
		fmt.Println("Received eventDate:", eventDate) // Log to see the format of the date received

		var parsedDate time.Time
		var err error

		// Try different date formats
		formats := []string{
			"2006-01-02T15:04",    // Format expected from datetime-local input
			"2006-01-02 15:04:00", // Format received in the log
			"2006-01-02 15:04",    // Without seconds
			"02/01/2006 15:04",    // Another possible format
		}

		for _, format := range formats {
			parsedDate, err = time.Parse(format, eventDate)
			if err == nil {
				break
			}
			fmt.Println("Error parsing date with format", format, ":", err)
		}

		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusBadRequest,
				Status:  "error",
				Message: "Format tanggal tidak valid. Format yang diharapkan adalah YYYY-MM-DDTHH:MM",
			}
			return c.Status(http.StatusBadRequest).JSON(response)
		}

		existingEvent.TanggalPelaksanaan = parsedDate
	}

	// Handle image file if provided
	file, err := c.FormFile("thumbnail_event")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal membuka file gambar",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		defer src.Close()

		imageURL, err := helper.UploadFileToCloudinary(src, file.Filename)
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal mengupload gambar",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		existingEvent.ThumbnailEvent = imageURL
	}

	// Update timestamp
	existingEvent.UpdatedAt = time.Now()

	// Save the updated event to the database
	if err := database.DB.Save(&existingEvent).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal mengupdate event",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Event berhasil diedit",
		Data:    existingEvent,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func DeleteEvent(c *fiber.Ctx) error {
	eventID := c.Params("id")
	var event models.Event
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "event not found",
		})
	}
	if err := database.DB.Delete(&event, eventID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete event",
		})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "event deleted successfully",
	})
}
