package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strconv"
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
		Status:        "Pending",
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
			Status:    "Pending",
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return false
		}
		Product.Availability = false
		if err := tx.Model(&Product).Where("id = ?", Product.ID).Update("availability", Product.Availability).Error; err != nil {
			tx.Rollback()
			return false
		}
	}

	if err := tx.Where("user_id = ? ", UserID).Delete(&CartItems).Error; err != nil {
		tx.Rollback()
		return false
	}

	// transaction ends
	tx.Commit()

	return true

}

func GetSellerOrders(c *gin.Context) {
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

	var orderItems []models.OrderItem
	var orderResponses []models.GetSellerOrdersResponse

	if err := database.DB.Preload("Product").
		Preload("Order").
		Preload("User").
		Where("seller_id = ?", sellerIDStr).Find(&orderItems).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "no orders found for this seller",
		})
		return
	}

	if len(orderItems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "no orders found for this seller",
		})
		return
	}

	for _, orderItem := range orderItems {
		orderResponses = append(orderResponses, models.GetSellerOrdersResponse{
			OrderItemID:   orderItem.OrderItemID,
			OrderID:       orderItem.OrderID,
			UserID:        orderItem.UserID,
			UserName:      orderItem.User.Name,
			ProductID:     orderItem.ProductID,
			ProductName:   orderItem.Product.Name,
			Description:   orderItem.Product.Description,
			Image:         orderItem.Product.Image,
			SellerID:      orderItem.SellerID,
			Price:         orderItem.Price,
			Status:        orderItem.Status,
			PaymentMethod: orderItem.Order.PaymentMethod,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   orderResponses,
	})
}

func SellerChangeOrderStatus(c *gin.Context) {
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

	var Request models.SellerChangeStatusRequest

	if err := c.ShouldBindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid input",
		})
		return
	}

	if !IsValidStatus(Request.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid order status",
		})
		return
	}

	var orderItem models.OrderItem

	if err := database.DB.Where("seller_id = ? AND order_item_id = ?", sellerIDStr, Request.OrderItemID).
		First(&orderItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "order item not found for this seller",
		})
		return
	}

	orderItem.Status = Request.Status
	if err := database.DB.Model(&orderItem).Where("order_item_id = ?", orderItem.OrderItemID).
		Update("status", orderItem.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update order status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "order status updated successfully",
		"data":    "Now status is  " + orderItem.Status,
	})
}

func AdminGetSellerOrderStatuses(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
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

	orderIDStr := c.Query("orderid")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid order ID",
		})
		return
	}

	var orderItems []models.OrderItem

	if err := database.DB.Where("order_id = ?", orderID).Preload("Seller").Find(&orderItems).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "no order items found for this order",
		})
		return
	}

	var sellerStatuses []models.GetSellerOrderStatusResponse
	for _, orderItem := range orderItems {
		sellerStatuses = append(sellerStatuses, models.GetSellerOrderStatusResponse{
			OrderItemID: orderItem.OrderItemID,
			ProductID:   orderItem.ProductID,
			SellerID:    orderItem.SellerID,
			SellerName:  orderItem.Seller.UserName,
			Status:      orderItem.Status,
			Price:       orderItem.Price,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   sellerStatuses,
	})
}

func AdminChangeOrderStatus(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
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

	var Request models.SellerChangeStatusRequest

	if err := c.ShouldBindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid input",
		})
		return
	}

	if !IsValidStatus(Request.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid order status",
		})
		return
	}

	var order models.Order

	if err := database.DB.Where("order_id = ?", Request.OrderItemID).
		First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "order not found",
		})
		return
	}

	order.Status = Request.Status
	if err := database.DB.Model(&order).Where("order_id = ?", order.OrderID).
		Update("status", order.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update order status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "order status updated successfully",
		"data":    "Now status is  " + order.Status,
	})
}

func UserCheckOrderStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	_, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	orderIDStr := c.Query("orderid")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid order ID",
		})
		return
	}

	var orderItems []models.Order

	if err := database.DB.Where("order_id = ?", orderID).Preload("Seller").Find(&orderItems).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "no order items found for this order",
		})
		return
	}

	var sellerStatuses []models.GetSellerOrderStatusResponse
	for _, orderItem := range orderItems {
		sellerStatuses = append(sellerStatuses, models.GetSellerOrderStatusResponse{
			OrderItemID: orderItem.OrderItemID,
			ProductID:   orderItem.ProductID,
			SellerID:    orderItem.SellerID,
			SellerName:  orderItem.Seller.UserName,
			Status:      orderItem.Status,
			Price:       orderItem.Price,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   sellerStatuses,
	})
}

func IsValidStatus(status string) bool {
	for _, validStatus := range models.StatusOptions {
		if status == validStatus {
			return true
		}
	}
	return false
}
