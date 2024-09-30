package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListAllSellers(c *gin.Context) {

	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	var sellerResponse []models.SellerResponse
	var sellers []models.Seller

	tx := database.DB.Find(&sellers)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	for _, seller := range sellers {
		sellerResponse = append(sellerResponse, models.SellerResponse{
			ID:          seller.ID,
			User:        seller.User,
			UserName:    seller.UserName,
			Description: seller.Description,
			IsVerified:  seller.IsVerified,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved user information",
		"data": gin.H{
			"users": sellerResponse,
		},
	})
}

func VerifySeller(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	sellerId := c.Query("sellerid")

	if sellerId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "userid is required",
		})
		return
	}

	var seller models.Seller

	if err := database.DB.First(&seller, sellerId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to fetch seller from the database",
		})
		return
	}
	if seller.IsVerified {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":  "failed",
			"message": "seller is already verified",
		})
		return
	}

	seller.IsVerified = true

	tx := database.DB.Model(&seller).Update("is_verified", seller.IsVerified)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to change the verification status ",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully verify the seller",
	})
}

func NotVerifySeller(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	sellerId := c.Query("sellerid")

	if sellerId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "userid is required",
		})
		return
	}

	var seller models.Seller

	if err := database.DB.First(&seller, sellerId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to fetch seller from the database",
		})
		return
	}
	if !seller.IsVerified {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":  "failed",
			"message": "seller is already Notverified",
		})
		return
	}

	seller.IsVerified = false

	tx := database.DB.Model(&seller).Update("is_verified", seller.IsVerified)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to change the verification status ",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "successfully verify the seller",
	})
}
