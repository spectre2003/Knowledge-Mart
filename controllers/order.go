package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func PlaceOrder(c *gin.Context) {
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

	var placeOrder models.PlaceOrder

	if err := c.BindJSON(&placeOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to bind the json",
		})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&placeOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var User models.User

	if err := database.DB.Where("id = ?", userIDStr).First(&User).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "user doesn't exist, please verify user id",
		})
		return
	}

	var CartItems []models.Cart
	var TotalAmount float64

	if err := database.DB.Preload("Product").Where("user_id = ?", userIDStr).Find(&CartItems).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to find the cart",
		})
		return
	}

	if len(CartItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Your cart is empty.",
		})
		return
	}

	for _, item := range CartItems {
		Product := item.Product
		if !Product.Availability {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "items in the cart are out of stock, please update the cart to ensure all items are in stock",
			})
			return
		}
		TotalAmount += Product.Price
	}

	var Address models.Address

	if err := database.DB.Where("user_id = ? AND id = ?", userIDStr, placeOrder.AddressID).First(&Address).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "failed",
			"message": "invalid address, please retry with user's address",
		})
		return
	}

	order := models.Order{
		UserID:        userIDStr,
		TotalAmount:   TotalAmount,
		PaymentMethod: placeOrder.PaymentMethod,
		PaymentStatus: "Pending",
		OrderedAt:     time.Now(),
		ShippingAddress: models.ShippingAddress{
			StreetName:   Address.StreetName,
			StreetNumber: Address.StreetNumber,
			City:         Address.City,
			State:        Address.State,
			PinCode:      Address.PinCode,
		},
	}

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to create order" + err.Error(),
		})
		return
	}

	if !CartToOrderItems(userIDStr, order) {
		database.DB.Delete(&order)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to transfer cart items to order",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order is successfully created",
		"data": gin.H{
			"order_details": order,
		},
	})
}

func CartToOrderItems(UserID uint, Order models.Order) bool {
	var CartItems []models.Cart

	if err := database.DB.Preload("Product").Where("user_id = ?", UserID).Find(&CartItems).Error; err != nil {
		return false
	}

	if len(CartItems) == 0 {
		return false
	}

	// transaction starts
	tx := database.DB.Begin()

	for _, cartItem := range CartItems {
		Product := cartItem.Product

		orderItem := models.OrderItem{
			OrderID:   Order.OrderID,
			ProductID: cartItem.ProductID,
			UserID:    UserID,
			SellerID:  Product.SellerID,
			Price:     Product.Price,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return false
		}
	}

	if err := tx.Where("user_id = ? AND order_id = ?", UserID, Order.OrderID).Delete(&CartItems).Error; err != nil {
		tx.Rollback()
		return false
	}

	// transaction ends
	tx.Commit()

	return true

}
