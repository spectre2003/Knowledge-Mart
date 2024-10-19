package controllers

import (
	"errors"
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ApplyReferralOnCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	UserIDStr, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	refCode := c.Query("refcode")

	if refCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "refrral code is required",
		})
		return
	}

	var CartItems []models.Cart
	if err := database.DB.Preload("Product").Where("user_id = ?", UserIDStr).Find(&CartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to fetch cart items. Please try again later.",
		})
		return
	}

	if len(CartItems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "Your cart is empty.",
		})
		return
	}

	var sum float64
	var CartResponse []models.CartResponse

	for _, item := range CartItems {
		var Product models.Product

		if err := database.DB.Preload("Seller").Where("id = ?", item.ProductID).First(&Product).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "Failed to fetch product information. Please try again later.",
			})
			return
		}

		CartResponse = append(CartResponse, models.CartResponse{
			ProductID:    item.ProductID,
			ProductName:  item.Product.Name,
			CategoryID:   item.Product.CategoryID,
			Description:  item.Product.Description,
			Price:        item.Product.Price,
			Availability: item.Product.Availability,
			Image:        item.Product.Image,
			SellerRating: Product.Seller.AverageRating,
			ID:           item.ID,
		})

		//ProductOfferAmount += float64(ProductOfferAmount) * float64()
		sum += Product.Price

	}

	var refDiscount float64
	var finalAmount float64

	var referredByUser models.User
	if err := database.DB.Where("referral_code = ?", refCode).First(&referredByUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "Invalid referral code. Please check and try again.",
		})
		return
	}

	var userReferralHistory models.UserReferralHistory

	if err := database.DB.Where("user_id = ?", userID).First(&userReferralHistory).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "database error occurred while checking referral history",
			})
			return
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "user already used referral",
		})
		return
	}

	refDiscount = float64(sum) * (5 / 100.0)
	finalAmount = sum - refDiscount

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"cart_items":      CartResponse,
			"total_amount":    fmt.Sprintf("%.2f", sum),
			"coupon_discount": fmt.Sprintf("%.2f", refDiscount),
			//"product_offer_amount": ProductOfferAmount,
			"final_amount": fmt.Sprintf("%.2f", finalAmount),
		},
		"message": "Cart items retrieved successfully",
	})
}

func ApplyReferral(totalAmount float64, userID uint, refCode string) (bool, string, float64) {
	var userReferralHistory models.UserReferralHistory

	if err := database.DB.Where("user_id = ?", userID).First(&userReferralHistory).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "database error occurred while checking referral history", 0
		}
	} else {
		return false, "user already used referral", 0
	}

	var referredByUser models.User
	if err := database.DB.Where("referral_code = ?", refCode).First(&referredByUser).Error; err != nil {
		return false, "invalid referral code", 0
	}

	discount := totalAmount * 0.05

	newReferralHistory := models.UserReferralHistory{
		UserID:       userID,
		ReferredBy:   referredByUser.ID,
		ReferralCode: refCode,
		ReferClaimed: true,
	}

	if err := database.DB.Create(&newReferralHistory).Error; err != nil {
		return false, "referral history creation failed", 0
	}

	return true, "", discount
}
