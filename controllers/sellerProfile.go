package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func GetSellerProfile(c *gin.Context) {
	var seller models.Seller

	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	sellerIDStr, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}

	if err := database.DB.Preload("User").Where("id = ? AND deleted_at IS NULL", sellerIDStr).First(&seller).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "seller not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved seller profile",
		"data": gin.H{
			"username_of_seller": seller.UserName,
			"seller_name":        seller.User.Name,
			"describtion":        seller.Description,
			"seller_id":          seller.ID,
			"average_rating":     seller.AverageRating,
			"wallet_amount":      seller.WalletAmount,
		},
	})
}

func EditSellerProfile(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	sellerIDStr, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}

	var Request models.EditSellerProfileRequest

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
		return
	}
	validate := validator.New()
	if err := validate.Struct(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	var existingSeller models.Seller

	if err := database.DB.Where("id = ?", sellerIDStr).First(&existingSeller).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to fetch seller details from the database",
		})
		return
	}

	if Request.UserName != "" {
		existingSeller.UserName = Request.UserName
	}
	if Request.Description != "" {
		existingSeller.Description = Request.Description
	}

	if err := database.DB.Model(&existingSeller).Updates(existingSeller).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update seller profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully updated seller information",
		"data": gin.H{
			"id":          existingSeller.ID,
			"username":    existingSeller.UserName,
			"description": existingSeller.Description,
		},
	})
}

func EditSellerPassword(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	sellerIDStr, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}

	var Request models.EditPasswordRequest

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	var existingSeller models.Seller

	if err := database.DB.Where("id = ?", sellerIDStr).First(&existingSeller).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to fetch user from the database",
		})
		return
	}
	err := CheckPassword(existingSeller.Password, Request.CurrentPassword)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Incorrect seller password",
		})
		return
	}
	if Request.NewPassword != Request.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "New password and confirm password do not match",
		})
		return
	}

	hashpassword, err := HashPassword(Request.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "error in password hashing" + err.Error(),
		})
		return
	}

	existingSeller.Password = hashpassword

	if err := database.DB.Model(&existingSeller).Select("password").Updates(existingSeller).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully updated user password",
	})
}
