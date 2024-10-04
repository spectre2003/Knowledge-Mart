package controllers

import (
	"fmt"
	"knowledgeMart/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized",
		})
		return
	}

	_, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if filepath.Ext(file.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only PDF files are allowed"})
		return
	}

	// Ensure temp directory exists
	tempDir := "./temp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err := os.Mkdir(tempDir, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create temp directory"})
			return
		}
	}

	// Save file temporarily
	tempFilePath := fmt.Sprintf("%s/%s", tempDir, file.Filename)
	err = c.SaveUploadedFile(file, tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save file temporarily"})
		return
	}

	// Upload to Cloudinary
	secureURL, err := utils.UploadFileToCloudinary(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file to cloud"})
		return
	}

	// Remove the temporary file after successful upload
	if err := os.Remove(tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clean up temp file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "file uploaded successfully",
		"file_url": secureURL,
	})
}
