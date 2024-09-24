package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func SellerRegister(c *gin.Context) {
	var Register models.SellerRegisterRequest
	var user models.User
	var newSeller models.Seller

	// Extract the userID from the context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "user not authorized",
		})
		return
	}

	_, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve user information",
		})
		return
	}

	// Bind JSON input to the SellerRegisterRequest struct
	if err := c.ShouldBindJSON(&Register); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process the incoming request: " + err.Error(),
		})
		return
	}
	Validate = validator.New()

	err := Validate.Struct(Register)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	// Check if the user exists
	if err := database.DB.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": "user not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve user information",
		})
		return
	}

	// Check if the seller already exists for the user
	var seller models.Seller
	tx := database.DB.Where("user_id = ? AND deleted_at IS NULL", userID).First(&seller)
	if tx.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "seller already exists for this user",
		})
		return
	} else if tx.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve seller information",
		})
		return
	}

	// Create the new seller associated with the user
	newSeller = models.Seller{
		UserID:      userID.(uint), // Use extracted userID
		Name:        Register.Name,
		Description: Register.Description,
		IsVerified:  false,
	}

	// Save the new seller
	if err := database.DB.Create(&newSeller).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to create a new seller",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Seller registered successfully",
		"data":    newSeller,
	})
}
