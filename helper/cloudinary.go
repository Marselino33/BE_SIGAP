package helper

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func UploadFileToCloudinary(file io.Reader, filename string) (string, error) {
	cloudName := os.Getenv("CLOUD_NAME")
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("API_SECRET")

	cldService, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return "", fmt.Errorf("failed to create Cloudinary service: %v", err)
	}
	ctx := context.Background()
	resp, err := cldService.Upload.Upload(ctx, file, uploader.UploadParams{})
	if err != nil {
		return "", fmt.Errorf("failed to upload image to Cloudinary: %v", err)
	}
	return resp.SecureURL, nil
}

func UploadMultipleFileToCloudinary(files []*multipart.FileHeader) ([]string, error) {
	cloudName := os.Getenv("CLOUD_NAME")
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("API_SECRET")

	cldService, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudinary service: %v", err)
	}

	var imageURLs []string
	ctx := context.Background()

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open image file: %v", err)
		}
		defer file.Close()

		resp, err := cldService.Upload.Upload(ctx, file, uploader.UploadParams{})
		if err != nil {
			return nil, fmt.Errorf("failed to upload image to Cloudinary: %v", err)
		}
		fmt.Println("Result image RESPONSE CKOUDINARY", resp)
		imageURLs = append(imageURLs, resp.URL)
	}
	fmt.Println("Result image urls dari cloduinary", imageURLs)
	return imageURLs, nil
}
