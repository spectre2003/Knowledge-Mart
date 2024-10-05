package controllers

import (
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListAllSellers(c *gin.Context) {

	// Check if admin is authorized
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

	// Preload the User data for each seller
	tx := database.DB.Preload("User").Find(&sellers)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	for _, seller := range sellers {
		sellerResponse = append(sellerResponse, models.SellerResponse{
			ID:           seller.ID,
			UserID:       seller.UserID,
			User:         seller.User.Name,
			Email:        seller.User.Email,
			PhoneNumber:  seller.User.PhoneNumber,
			UserName:     seller.UserName,
			Description:  seller.Description,
			IsVerified:   seller.IsVerified,
			SellerRating: seller.AverageRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved seller information",
		"data": gin.H{
			"sellers": sellerResponse,
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

func SellerRating(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	var req models.RatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid request data",
		})
		return
	}

	var seller models.Seller
	if err := database.DB.Where("id = ?", req.SellerID).First(&seller).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "seller not found",
		})
		return
	}

	newRating := models.SellerRating{
		UserID:   userIDUint,
		SellerID: req.SellerID,
		Rating:   req.Rating,
	}
	if err := database.DB.Create(&newRating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to save rating: " + err.Error(),
		})
		return
	}

	if err := UpdateSellerAverageRating(req.SellerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update seller average rating: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Rating successfully submitted and seller rating updated",
		"data": gin.H{
			"user_id":   userIDUint,
			"seller_id": req.SellerID,
			"rating":    req.Rating,
		},
	})
}

func UpdateSellerAverageRating(sellerID uint) error {
	var totalRatingSum float64
	var ratingCount int64

	if err := database.DB.Model(&models.SellerRating{}).
		Where("seller_id = ?", sellerID).
		Select("COALESCE(SUM(rating), 0)").Scan(&totalRatingSum).Error; err != nil {
		return fmt.Errorf("failed to calculate total rating sum: %w", err)
	}

	if err := database.DB.Model(&models.SellerRating{}).
		Where("seller_id = ?", sellerID).
		Count(&ratingCount).Error; err != nil {
		return fmt.Errorf("failed to count total ratings: %w", err)
	}

	if ratingCount == 0 {
		return nil
	}

	averageRating := totalRatingSum / float64(ratingCount)

	if err := database.DB.Model(&models.Seller{}).
		Where("id = ?", sellerID).
		Update("average_rating", averageRating).Error; err != nil {
		return fmt.Errorf("failed to update seller average rating: %w", err)
	}

	return nil
}
