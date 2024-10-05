package utils

import (
	"context"
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

func UploadFileToCloudinary(filePath string) (string, error) {
	cld, err := InitCloudinary()
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	// Make sure to set ResourceType to "raw"
	uploadResult, err := cld.Upload.Upload(ctx, filePath, uploader.UploadParams{
		ResourceType: "raw", // This is important for PDF and non-image files
	})
	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}
