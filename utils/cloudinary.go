package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func InitCloudinary() (*cloudinary.Cloudinary, error) {
	cloudinaryURL := os.Getenv("CLOUDINARYURL")
	if cloudinaryURL == "" {
		log.Fatal("CLOUDINARYURL environment variable not set")
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, err
	}
	return cld, nil
}

func UploadFileToCloudinary(filePath string, resourceType string) (string, error) {
	cloudName := os.Getenv("CLOUDNAME")
	apiKey := os.Getenv("CLOUDINARYACCESSKEY")
	apiSecret := os.Getenv("CLOUDINARYSECRETKEY")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return "", fmt.Errorf("cloudinary configuration is missing in environment variables")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return "", fmt.Errorf("failed to create cloudinary instance: %w", err)
	}

	uploadResult, err := cld.Upload.Upload(context.TODO(), filePath, uploader.UploadParams{
		ResourceType: resourceType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to cloudinary: %w", err)
	}

	return uploadResult.SecureURL, nil
}
