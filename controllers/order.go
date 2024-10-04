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
	var sellerID uint

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

		if sellerID == 0 {
			sellerID = Product.SellerID
		} else if sellerID != Product.SellerID {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "You can only add products from one seller to your cart per order.",
			})
			return
		}
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
		PaymentStatus: "pending",
		OrderedAt:     time.Now(),
		SellerID:      sellerID,
		Status:        "pending",
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

func GetUserOrders(c *gin.Context) {
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

	var orders []models.Order
	var orderResponses []models.GetSellerOrdersResponse

	if err := database.DB.Where("seller_id = ?", sellerIDStr).Find(&orders).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "no orders found for this seller",
		})
		return
	}

	for _, order := range orders {
		var orderItems []models.OrderItem
		if err := database.DB.Where("order_id = ?", order.OrderID).Find(&orderItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve order items",
			})
			return
		}

		var products []models.ProductArray
		for _, item := range orderItems {
			var product models.Product
			if err := database.DB.Where("id = ?", item.ProductID).First(&product).Error; err == nil {
				products = append(products, models.ProductArray{
					ProductID:   product.ID,
					ProductName: product.Name,
					Description: product.Description,
					Image:       product.Image,
					Price:       item.Price,
					OrderItemID: item.OrderItemID,
				})
			}
		}
		orderResponses = append(orderResponses, models.GetSellerOrdersResponse{
			OrderID:         order.OrderID,
			UserID:          order.UserID,
			SellerID:        order.SellerID,
			PaymentMethod:   order.PaymentMethod,
			TotalAmount:     order.TotalAmount,
			OrderStatus:     order.Status,
			Product:         products,
			ShippingAddress: order.ShippingAddress,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   orderResponses,
	})
}
func SellerUpdateOrderStatus(c *gin.Context) {
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

	orderId := c.Query("orderid")

	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "orderId is required",
		})
		return
	}

	var orders models.Order

	if err := database.DB.Where("seller_id = ? AND order_id = ?", sellerIDStr, orderId).
		First(&orders).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "orders not found for this seller",
		})
		return
	}

	switch orders.Status {
	case models.OrderStatusPending:
		orders.Status = models.OrderStatusShipped
	case models.OrderStatusShipped:
		orders.Status = models.OrderStatusOutForDelivery
	case models.OrderStatusOutForDelivery:
		orders.Status = models.OrderStatusDelivered
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid order status transition",
		})
		return
	}

	if orders.Status == models.OrderStatusDelivered {
		orders.PaymentStatus = models.PaymentStatusPaid
	}

	if err := database.DB.Model(&orders).Updates(map[string]interface{}{
		"status":         orders.Status,
		"payment_status": orders.PaymentStatus,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update order status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "order status updated successfully",
		"data": gin.H{
			"newStatus": orders.Status,
		},
	})
}

func UserCheckOrderStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized",
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

	var orders []models.Order

	if err := database.DB.Where("user_id = ?", userIDStr).Find(&orders).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "no orders found for this user",
		})
		return
	}

	var userOrderResponses []models.UserOrderResponse
	for _, order := range orders {
		var orderItems []models.OrderItem
		if err := database.DB.Preload("Product").Preload("Seller").Where("order_id = ?", order.OrderID).Find(&orderItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve order items",
			})
			return
		}

		var orderItemResponses []models.OrderItemResponse
		for _, orderItem := range orderItems {
			orderItemResponses = append(orderItemResponses, models.OrderItemResponse{
				OrderItemID: orderItem.OrderItemID,
				ProductName: orderItem.Product.Name,
				CategoryID:  orderItem.Product.CategoryID,
				Description: orderItem.Product.Description,
				Price:       orderItem.Price,
				Image:       orderItem.Product.Image,
				SellerName:  orderItem.Seller.UserName,
			})
		}

		userOrderResponses = append(userOrderResponses, models.UserOrderResponse{
			OrderID:         order.OrderID,
			OrderedAt:       order.OrderedAt,
			TotalAmount:     order.TotalAmount,
			Items:           orderItemResponses,
			Status:          order.Status,
			ShippingAddress: order.ShippingAddress,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   userOrderResponses,
	})
}

func CancelOrder(c *gin.Context) {
	sellerID, isSeller := c.Get("sellerID")
	userID, isUser := c.Get("userID")

	if !isSeller && !isUser {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user or seller not authorized",
		})
		return
	}

	var id uint
	if isSeller {
		id = sellerID.(uint)
	} else if isUser {
		id = userID.(uint)
	}

	orderId := c.Query("orderid")

	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "orderId is required",
		})
		return
	}

	var orders models.Order
	var condition string
	if isSeller {
		condition = "seller_id = ? AND order_id = ?"
	} else {
		condition = "user_id = ? AND order_id = ?"
	}

	if err := database.DB.Where(condition, id, orderId).First(&orders).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "order not found for this user or seller",
		})
		return
	}

	orders.Status = models.OrderStatusCanceled

	if err := database.DB.Model(&orders).Update("status", orders.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update order status",
		})
		return
	}

	var orderItems []models.OrderItem

	if err := database.DB.Preload("Product").Where("order_id = ?", orderId).Find(&orderItems).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "order items not found for this order",
		})
		return
	}

	for _, orderItem := range orderItems {
		orderItem.Product.Availability = true

		if err := database.DB.Model(&orderItem.Product).Update("availability", orderItem.Product.Availability).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to update product availability",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order cancelled successfully",
	})
}
