package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"knowledgeMart/utils"
	"net/http"
)

func SellerRegister(c *gin.Context) {
	var Register models.SellerRegisterRequest
	var user models.User
	var newSeller models.Seller

	// Extract the userID from the context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	fmt.Println("Extracted userID: ", userID)

	userIDStr, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	// Bind JSON input to the SellerRegisterRequest struct
	if err := c.ShouldBindJSON(&Register); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process the incoming request: " + err.Error(),
		})
		return
	}
	Validate = validator.New()

	err := Validate.Struct(Register)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	// Check if the user exists
	if err := database.DB.Where("id = ? AND deleted_at IS NULL", userIDStr).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "user not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	// Check if the seller already exists for the user
	var seller models.Seller
	tx := database.DB.Where("user_id = ? AND deleted_at IS NULL", userIDStr).First(&seller)
	if tx.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "seller already exists for this user",
		})
		return
	} else if tx.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}
	hashpassword, err := HashPassword(Register.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "error in password hashing" + err.Error(),
		})
		return
	}

	// Create the new seller associated with the user
	newSeller = models.Seller{
		UserID:      userID.(uint),
		UserName:    Register.UserName,
		Password:    hashpassword,
		Description: Register.Description,
		IsVerified:  false,
	}

	if err := database.DB.Create(&newSeller).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to create a new seller",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Seller registered successfully ,status is pending , please login to continue",
		"data": gin.H{
			"seller_name":     user.Name,
			"seller_username": newSeller.UserName,
			"description":     newSeller.Description,
			"eamil":           user.Email,
			"phone_number":    user.PhoneNumber,
			"user_id":         newSeller.UserID,
			"sellerId":        newSeller.ID,
		},
	})
}

func SellerLogin(c *gin.Context) {
	var LoginSeller models.SellerLoginRequest

	if err := c.ShouldBindJSON(&LoginSeller); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	Validate = validator.New()

	err := Validate.Struct(LoginSeller)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var seller models.Seller
	tx := database.DB.Where("user_name = ? AND deleted_at is NULL", LoginSeller.UserName).First(&seller)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid email or password",
		})
		return
	}
	err = CheckPassword(seller.Password, LoginSeller.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Incorrect password",
		})
		return
	}
	if !seller.IsVerified {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller is not verified , status is pending ",
		})
		return
	}

	token, err := utils.GenerateJWT(seller.ID, "seller")
	if token == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Seller Login successfully",
		"data": gin.H{
			"token":    token,
			"username": seller.UserName,
			"verified": seller.IsVerified,
		},
	})
}
