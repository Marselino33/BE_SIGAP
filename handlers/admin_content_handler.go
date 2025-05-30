package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetAllContents(c *fiber.Ctx) error {
	var contents []models.Content
	if err := database.DB.Preload("ViolenceCategory").Find(&contents).Error; err != nil {
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
		Data:    contents,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func GetContentByID(c *fiber.Ctx) error {
	contentID := c.Params("id")

	var content models.Content
	if err := database.DB.Preload("ViolenceCategory").First(&content, contentID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Content not found",
		})
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Content details",
		Data:    content,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func CreateContent(c *fiber.Ctx) error {
	var content models.Content
	if err := c.BodyParser(&content); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	file, err := c.FormFile("image_content")
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
			Message: "Gagal Mengupload Gambar",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	content.ImageContent = imageURL
	content.Judul = c.FormValue("judul")
	content.IsiContent = c.FormValue("isi_content")
	violenceCategoryID, err := strconv.ParseInt(c.FormValue("violence_category_id"), 10, 64)
	if err != nil || violenceCategoryID == 0 {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Id Kategory Kekerasan Tidak Ditemukan ",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	var violenceCategory models.ViolenceCategory
	if err := database.DB.First(&violenceCategory, violenceCategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response := helper.ResponseWithOutData{
				Code:    http.StatusBadRequest,
				Status:  "error",
				Message: "Kategori Kekerasan Tidak ditemukan",
			}
			return c.Status(http.StatusBadRequest).JSON(response)
		}
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to check violence category",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	content.ViolenceCategoryID = uint(violenceCategoryID)
	if err := database.DB.Create(&content).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal Membuat Konten",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	if err := database.DB.Preload("ViolenceCategory").First(&content, content.ID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to load content with violence category",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	response := helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Berhasil Membuat Konten",
		Data:    content,
	}
	return c.Status(http.StatusCreated).JSON(response)
}

func UpdateContent(c *fiber.Ctx) error {
	contentID := c.Params("id")

	// Fetch existing content from the database
	var existingContent models.Content
	if err := database.DB.Preload("ViolenceCategory").First(&existingContent, contentID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Konten Tidak ada",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}

	// Parse form data
	judul := c.FormValue("judul")
	if judul != "" {
		existingContent.Judul = judul
	}

	isiContent := c.FormValue("isi_content")
	if isiContent != "" {
		existingContent.IsiContent = isiContent
	}

	violenceCategoryID := c.FormValue("violence_category_id")
	if violenceCategoryID != "" {
		vcID, err := strconv.ParseUint(violenceCategoryID, 10, 64)
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusBadRequest,
				Status:  "error",
				Message: "Kategori ID Salah",
			}
			return c.Status(http.StatusBadRequest).JSON(response)
		}

		// Check if the violence category exists in the database
		var violenceCategory models.ViolenceCategory
		if err := database.DB.First(&violenceCategory, vcID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response := helper.ResponseWithOutData{
					Code:    http.StatusBadRequest,
					Status:  "error",
					Message: "Tidak Dapat Mencari Kategori kekerasan",
				}
				return c.Status(http.StatusBadRequest).JSON(response)
			}
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal Memeriksa Kategori kekerasan",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		existingContent.ViolenceCategoryID = uint(vcID)
	}

	// Handle image file if provided
	file, err := c.FormFile("image_content")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal Membuka File Gambar",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		defer src.Close()

		imageURL, err := helper.UploadFileToCloudinary(src, file.Filename)
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal Mengupload Gambar",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		existingContent.ImageContent = imageURL
	}

	// Update timestamp
	existingContent.UpdatedAt = time.Now()

	// Save updated content to the database
	if err := database.DB.Save(&existingContent).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal Mengupdate Konten",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	// Reload content with related ViolenceCategory to include in response
	if err := database.DB.Preload("ViolenceCategory").First(&existingContent, contentID).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal Memuat Konten dengan Kategori Kekerasan",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Berhasil Mengupdate Konten",
		Data:    existingContent,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func DeleteContent(c *fiber.Ctx) error {
	contentID := c.Params("id")
	var content models.Content
	if err := database.DB.First(&content, contentID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Content not found",
		})
	}
	if err := database.DB.Delete(&content, contentID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete content",
		})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Kontent Berhasil Dihapus",
	})
}
