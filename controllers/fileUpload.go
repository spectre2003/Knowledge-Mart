package controllers

import (
	"knowledgeMart/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "User not authorized"})
		return
	}

	_, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to retrieve user information"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "File is required"})
		return
	}

	if filepath.Ext(file.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Only PDF files are allowed"})
		return
	}

	tempDir := "./temp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err := os.Mkdir(tempDir, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Unable to create temp directory"})
			return
		}
	}
	tempFilePath := filepath.Join(tempDir, file.Filename) // Use Join for better path handling
	err = c.SaveUploadedFile(file, tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Unable to save file temporarily"})
		return
	}

	secureURL, err := utils.UploadFileToCloudinary(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to upload file to cloud"})
		return
	}

	// Clean up the temporary file
	if err := os.Remove(tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to clean up temp file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "File uploaded successfully",
		"file_url": secureURL,
	})
}
