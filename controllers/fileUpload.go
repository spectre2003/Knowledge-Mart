package controllers

import (
	"knowledgeMart/utils"
	"log"
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
			"message": "User not authorized",
		})
		return
	}

	_, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve user information",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "File is required",
			"error":   err.Error(),
		})
		return
	}

	if filepath.Ext(file.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Only PDF files are allowed",
		})
		return
	}

	tempDir := "./temp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err := os.Mkdir(tempDir, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Unable to create temp directory",
				"error":   err.Error(),
			})
			return
		}
	}

	tempFilePath := filepath.Join(tempDir, file.Filename)
	err = c.SaveUploadedFile(file, tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Unable to save file temporarily",
			"error":   err.Error(),
		})
		return
	}

	if _, err := os.Stat(tempFilePath); os.IsNotExist(err) {
		log.Println("Temporary file does not exist:", tempFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Temporary file not found",
		})
		return
	}

	secureURL, err := utils.UploadFileToCloudinary(tempFilePath, "raw")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to upload file to cloud",
			"error":   err.Error(),
		})
		return
	}

	log.Println("File uploaded to Cloudinary:", secureURL)

	if err := os.Remove(tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to clean up temp file",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "File uploaded successfully",
		"file_url": secureURL,
	})
}
