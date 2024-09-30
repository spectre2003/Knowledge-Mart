package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func AddToCart(c *gin.Context) {
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

	var Request models.AddToCartRequest
	var Product models.Product
	var CartItem models.CartItems
	var UserCart models.UserCart

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

	if err := database.DB.Where("id = ?", Request.ProductID).First(&Product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Failed to fetch product information. Please ensure the specified product exists.",
		})
		return
	}

	if !Product.Availability {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Product is not available",
		})
		return
	}

	UserCart.UserID = UserIDStr

	if err := database.DB.Create(&UserCart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fails",
			"message": "Failed to update cart items. Please try again later.",
		})
		return
	}

	CartItem.UserCartID = UserCart.ID
	CartItem.ProductID = Product.ID

	if err := database.DB.Create(&CartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fails",
			"message": "Failed to update cart items. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Product added to cart successfully",
	})

}

func ListAllCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
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

	var user models.User
	if err := database.DB.Where("id = ?", userIDUint).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	var Cart []models.CartItems

	tx := database.DB.Select("*").Find(&Cart)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the database",
		})
		return
	}
}
