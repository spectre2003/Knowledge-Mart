package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strconv"

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

	var Product models.Product

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

	var existingCartItem models.Cart

	if err := database.DB.Where("product_id = ? AND user_id = ?", Request.ProductID, UserIDStr).First(&existingCartItem).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Product is already in your cart.",
		})
		return
	}

	cart := models.Cart{
		ProductID: Request.ProductID,
		UserID:    UserIDStr,
	}

	if err := database.DB.Create(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fails",
			"message": "Failed to add product to cart. Please try again later.",
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

	var Carts []models.Cart

	if err := database.DB.Preload("Product").Where("user_id = ?", userIDUint).Find(&Carts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve cart information",
		})
		return
	}

	if len(Carts) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Your cart is empty.",
		})
		return
	}

	var CartResponse []models.CartResponse
	var TotalAmount float64

	for _, cart := range Carts {
		TotalAmount += cart.Product.Price
		CartResponse = append(CartResponse, models.CartResponse{
			ProductID:    cart.ProductID,
			ProductName:  cart.Product.Name,
			CategoryID:   cart.Product.CategoryID,
			Description:  cart.Product.Description,
			Price:        cart.Product.Price,
			Availability: cart.Product.Availability,
			Image:        cart.Product.Image,
			CartID:       cart.ID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully get all the cart items",
		"data": gin.H{
			"Cart":         CartResponse,
			"Total Amount": TotalAmount,
		},
	})
}

func RemoveItemFromCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	userIDStr, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	cartIDStr := c.Query("cartid")
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid cartID",
		})
		return
	}

	var cart models.Cart

	if err := database.DB.First(&cart, cartID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "product is not present in the cart",
		})
		return
	}

	if cart.UserID != userIDStr {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "You are not authorized to remove this item from the cart.",
		})
		return
	}

	if err := database.DB.Delete(&cart, cartID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "unable to remove product from the cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully remove product from cart",
	})
}
