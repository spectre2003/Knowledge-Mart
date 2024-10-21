package controllers

import (
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func RoundDecimalValue(value float64) float64 {
	multiplier := math.Pow(10, 2)
	return math.Round(value*multiplier) / multiplier
}

func PlaceOrder(c *gin.Context) {
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

	var request models.PlaceOrder
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&request); err != nil {
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
			"message": "user doesn't exist, please verify user ID",
		})
		return
	}

	var CartItems []models.Cart
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

	var discountedPrice float64
	var CategoryDiscount float64
	var TotalAmount float64
	var finalAmount float64
	var sellerID uint

	for _, item := range CartItems {
		Product := item.Product

		if !Product.Availability {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "Some items in the cart are out of stock.",
			})
			return
		}

		var category models.Category
		if err := database.DB.First(&category, Product.CategoryID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "failed to find category for the product",
			})
			return
		}

		discountedPrice = calculateFinalAmount(Product.Price, Product.OfferAmount, category.OfferPercentage)

		TotalAmount += Product.Price
		finalAmount += discountedPrice

		categoryDiscount := Product.Price - discountedPrice
		CategoryDiscount += categoryDiscount

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

	// Apply coupon discount if available
	var CouponDiscount float64
	if request.CouponCode != "" {
		success, msg, discount := ApplyCouponToOrder(TotalAmount, userIDStr, request.CouponCode)
		if !success {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": msg,
			})
			return
		}
		CouponDiscount = discount
	}
	finalAmount -= CouponDiscount

	var ReferralDiscount float64
	if request.ReferralCode != "" {
		success, msg, discount := ApplyReferral(finalAmount, userIDStr, request.ReferralCode)
		if !success {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": msg,
			})
			return
		}
		ReferralDiscount = discount
	}

	finalAmount -= ReferralDiscount

	var Address models.Address
	if err := database.DB.Where("user_id = ? AND id = ?", userIDStr, request.AddressID).First(&Address).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "failed",
			"message": "Invalid shipping address.",
		})
		return
	}

	PaymentMethodOption := ""
	switch request.PaymentMethod {
	case 1:
		PaymentMethodOption = models.Razorpay
	case 2:
		PaymentMethodOption = models.Wallet
	case 3:
		PaymentMethodOption = models.COD
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid payment method.",
		})
		return
	}

	status := models.OrderStatusPending

	if PaymentMethodOption == models.COD {
		status = models.OrderStatusConfirmed
	}

	tx := database.DB.Begin()

	order := models.Order{
		UserID:                 userIDStr,
		TotalAmount:            TotalAmount,
		FinalAmount:            RoundDecimalValue(finalAmount),
		PaymentMethod:          PaymentMethodOption,
		PaymentStatus:          models.OrderStatusPending,
		OrderedAt:              time.Now(),
		CouponCode:             request.CouponCode,
		CouponDiscountAmount:   RoundDecimalValue(CouponDiscount),
		ReferralDiscountAmount: RoundDecimalValue(ReferralDiscount),
		CategoryDiscountAmount: RoundDecimalValue(CategoryDiscount),
		SellerID:               sellerID,
		Status:                 status,
		ShippingAddress: models.ShippingAddress{
			StreetName:   Address.StreetName,
			StreetNumber: Address.StreetNumber,
			City:         Address.City,
			State:        Address.State,
			PinCode:      Address.PinCode,
			PhoneNumber:  Address.PhoneNumber,
		},
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to create order.",
		})
		return
	}

	if PaymentMethodOption == models.Wallet {
		orderIDStr := fmt.Sprintf("%d", order.OrderID)
		if _, err := ProcessWalletPayment(userIDStr, orderIDStr, CouponDiscount, ReferralDiscount, tx); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}
	}

	if PaymentMethodOption == models.COD {
		if !CartToOrderItems(userIDStr, order, CouponDiscount, ReferralDiscount) {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Failed to transfer cart items to order.",
			})
			return
		}
	}

	// Commit the transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order successfully created with " + order.PaymentMethod,
		"data": gin.H{
			"order_id":      order.OrderID,
			"order_details": order,
		},
	})
}

func CartToOrderItems(UserID uint, Order models.Order, CouponDiscount float64, ReferralDiscount float64) bool {
	var CartItems []models.Cart
	if err := database.DB.Preload("Product").Where("user_id = ?", UserID).Find(&CartItems).Error; err != nil {
		return false
	}

	if len(CartItems) == 0 {
		return false
	}

	var totalCartPrice float64
	for _, cartItem := range CartItems {
		totalCartPrice += cartItem.Product.OfferAmount
	}

	totalDiscount := CouponDiscount + ReferralDiscount

	tx := database.DB.Begin()

	for _, cartItem := range CartItems {
		Product := cartItem.Product

		var category models.Category
		if err := database.DB.First(&category, Product.CategoryID).Error; err != nil {
			tx.Rollback()
			return false
		}

		discountedPrice := calculateFinalAmount(Product.Price, Product.OfferAmount, category.OfferPercentage)

		finalPrice := discountedPrice
		if totalDiscount > 0 {
			proportionalDiscount := (Product.OfferAmount / totalCartPrice) * totalDiscount
			finalPrice -= proportionalDiscount
		}

		fmt.Println(finalPrice)

		orderItem := models.OrderItem{
			OrderID:            Order.OrderID,
			ProductID:          cartItem.ProductID,
			UserID:             UserID,
			SellerID:           Product.SellerID,
			Price:              finalPrice,
			ProductOfferAmount: Product.OfferAmount,
			Status:             models.OrderStatusConfirmed,
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

	if err := tx.Where("user_id = ?", UserID).Delete(&CartItems).Error; err != nil {
		tx.Rollback()
		return false
	}

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
		if err := database.DB.Preload("Product").Where("order_id = ?", order.OrderID).Find(&orderItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve order items",
			})
			return
		}

		var orderItemResponses []models.OrderItemResponse
		for _, item := range orderItems {
			orderItemResponses = append(orderItemResponses, models.OrderItemResponse{
				OrderItemID: item.OrderItemID,
				ProductName: item.Product.Name,
				CategoryID:  item.Product.CategoryID,
				Description: item.Product.Description,
				Price:       item.Price,
				Image:       item.Product.Image,
				//SellerName:  item.Seller.UserName,
				OrderStatus: item.Status,
			})
		}

		orderResponses = append(orderResponses, models.GetSellerOrdersResponse{
			OrderID:         order.OrderID,
			UserID:          order.UserID,
			SellerID:        order.SellerID,
			PaymentMethod:   order.PaymentMethod,
			PaymentStatus:   order.PaymentStatus,
			TotalAmount:     RoundDecimalValue(order.TotalAmount),
			FinalAmount:     RoundDecimalValue(order.FinalAmount),
			OrderStatus:     order.Status,
			Product:         orderItemResponses,
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

	orderItemId := c.Query("orderitemid")
	if orderItemId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "orderItemId is required",
		})
		return
	}

	var ordersItem models.OrderItem
	if err := database.DB.Where("seller_id = ? AND order_item_id = ?", sellerIDStr, orderItemId).
		First(&ordersItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "ordersItem not found for this seller",
		})
		return
	}

	var order models.Order
	if err := database.DB.Where("order_id = ?", ordersItem.OrderID).Find(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "order not found",
		})
		return
	}

	switch ordersItem.Status {
	case models.OrderStatusPending:
		ordersItem.Status = models.OrderStatusConfirmed
	case models.OrderStatusConfirmed:
		ordersItem.Status = models.OrderStatusShipped
	case models.OrderStatusShipped:
		ordersItem.Status = models.OrderStatusOutForDelivery
	case models.OrderStatusOutForDelivery:
		ordersItem.Status = models.OrderStatusDelivered
	case models.OrderStatusDelivered:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Order already delivered",
		})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid order status transition",
		})
		return
	}

	tx := database.DB.Begin()

	if err := tx.Model(&ordersItem).Updates(map[string]interface{}{
		"status": ordersItem.Status,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to update order item status",
		})
		return
	}

	var orderItems []models.OrderItem
	if err := tx.Where("order_id = ?", ordersItem.OrderID).Find(&orderItems).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve order items",
		})
		return
	}

	allDelivered := true
	allPending := true
	allConfirmed := true
	allOutforDelivery := true
	for _, item := range orderItems {
		if item.Status != models.OrderStatusDelivered {
			allDelivered = false
		}
		if item.Status != models.OrderStatusPending {
			allPending = false
		}
		if item.Status != models.OrderStatusConfirmed {
			allConfirmed = false
		}
		if item.Status != models.OrderStatusOutForDelivery {
			allOutforDelivery = false
		}
	}

	if allDelivered {
		order.Status = models.OrderStatusDelivered
	} else if allPending {
		order.Status = models.OrderStatusPending
	} else if allConfirmed {
		order.Status = models.OrderStatusConfirmed
	} else if allOutforDelivery {
		order.Status = models.OrderStatusOutForDelivery
	} else {
		order.Status = models.OrderStatusShipped
	}

	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":         order.Status,
		"payment_status": order.PaymentStatus,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to update overall order status",
		})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order status updated successfully",
		"data": gin.H{
			"newStatus": ordersItem.Status,
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
		if err := database.DB.Preload("Product").Where("order_id = ?", order.OrderID).Find(&orderItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve order items",
			})
			return
		}

		countPending, countShipped, countDelivered, countOutForDelivery, countConfirmed := 0, 0, 0, 0, 0

		var orderItemResponses []models.OrderItemResponse
		for _, orderItem := range orderItems {
			switch orderItem.Status {
			case models.OrderStatusPending:
				countPending++
			case models.OrderStatusShipped:
				countShipped++
			case models.OrderStatusDelivered:
				countDelivered++
			case models.OrderStatusOutForDelivery:
				countOutForDelivery++
			case models.OrderStatusConfirmed:
				countConfirmed++
			}

			orderItemResponses = append(orderItemResponses, models.OrderItemResponse{
				OrderItemID: orderItem.OrderItemID,
				ProductName: orderItem.Product.Name,
				CategoryID:  orderItem.Product.CategoryID,
				Description: orderItem.Product.Description,
				Price:       orderItem.Price,
				Image:       orderItem.Product.Image,
				//SellerName:  orderItem.Seller.UserName,
				OrderStatus: orderItem.Status,
			})
		}

		statusCounts := gin.H{}
		if countPending > 0 {
			statusCounts["Pending"] = countPending
		}
		if countShipped > 0 {
			statusCounts["Shipped"] = countShipped
		}
		if countDelivered > 0 {
			statusCounts["Delivered"] = countDelivered
		}
		if countOutForDelivery > 0 {
			statusCounts["OutForDelivery"] = countOutForDelivery
		}
		if countConfirmed > 0 {
			statusCounts["Confirmed"] = countConfirmed
		}

		userOrderResponses = append(userOrderResponses, models.UserOrderResponse{
			OrderID:         order.OrderID,
			OrderedAt:       order.OrderedAt,
			TotalAmount:     RoundDecimalValue(order.TotalAmount),
			FinalAmount:     RoundDecimalValue(order.FinalAmount),
			Items:           orderItemResponses,
			Status:          order.Status,
			PaymentStatus:   order.PaymentStatus,
			ShippingAddress: order.ShippingAddress,
			ItemCounts:      statusCounts,
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
	itemId := c.Query("itemid")

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

	tx := database.DB.Begin()

	// Cancel single item
	if itemId != "" {
		var orderItem models.OrderItem
		if err := tx.Where("order_id = ? AND order_item_id = ?", orderId, itemId).Preload("Product").First(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "order item not found",
			})
			return
		}

		orders.FinalAmount -= orderItem.Price
		if err := tx.Model(&orders).Update("final_amount", orders.TotalAmount).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to update order total",
			})
			return
		}

		orderItem.Product.Availability = true
		if err := tx.Model(&orderItem.Product).Update("availability", orderItem.Product.Availability).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to update product availability",
			})
			return
		}

		if orders.PaymentStatus == models.PaymentStatusPaid {
			err := RefundToUser(tx, id, orderId, orderItem.Price, "Single item canceled", isSeller)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "failed",
					"message": "failed to refund amount",
				})
				return
			}
		}

		orderItem.Status = models.OrderStatusCanceled
		if err := tx.Model(&orderItem).Update("status", orderItem.Status).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to cancel order item",
			})
			return
		}

		tx.Commit()

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Order item canceled and amount refunded",
		})
		return
	}

	orders.Status = models.OrderStatusCanceled
	orders.PaymentStatus = models.PaymentStatusCanceled

	if err := tx.Model(&orders).Updates(map[string]interface{}{
		"status":         orders.Status,
		"payment_status": orders.PaymentStatus,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update order status",
		})
		return
	}

	var orderItems []models.OrderItem
	if err := tx.Preload("Product").Where("order_id = ?", orderId).Find(&orderItems).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "order items not found for this order",
		})
		return
	}

	if orders.PaymentStatus == models.PaymentStatusPaid {
		err := RefundToUser(tx, id, orderId, orders.FinalAmount, "Entire order canceled", isSeller)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to refund amount " + err.Error(),
			})
			return
		}
	}

	for _, orderItem := range orderItems {
		orderItem.Status = models.OrderStatusCanceled
		if err := tx.Model(&orderItem).Update("status", orderItem.Status).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to update order item status",
			})
			return
		}

		orderItem.Product.Availability = true
		if err := tx.Model(&orderItem.Product).Update("availability", orderItem.Product.Availability).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to update product availability",
			})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order canceled and amount refunded",
	})
}

func ReturnOrder(c *gin.Context) {
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

	orderId := c.Query("orderid")
	itemId := c.Query("itemid")

	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "orderId is required",
		})
		return
	}

	var orders models.Order
	if err := database.DB.Where("user_id = ? AND order_id = ?", userIDStr, orderId).First(&orders).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "order not found for this user",
		})
		return
	}

	tx := database.DB.Begin()

	isSeller := false

	// Return single item
	if itemId != "" {
		var orderItem models.OrderItem
		if err := tx.Where("order_id = ? AND order_item_id = ?", orderId, itemId).Preload("Product").First(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "order item not found",
			})
			return
		}

		orderItem.Status = models.OrderStatusReturned
		if err := tx.Model(&orderItem).Update("status", orderItem.Status).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to return order item",
			})
			return
		}

		orders.FinalAmount -= orderItem.Price
		if err := tx.Model(&orders).Update("final_amount", orders.TotalAmount).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to update order total",
			})
			return
		}

		orderItem.Product.Availability = true
		if err := tx.Model(&orderItem.Product).Update("availability", orderItem.Product.Availability).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to update product availability",
			})
			return
		}

		if orders.PaymentStatus == models.PaymentStatusPaid {
			err := RefundToUser(tx, userIDStr, orderId, orderItem.Price, "Single item returned", isSeller)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "failed",
					"message": "failed to refund amount",
				})
				return
			}
		}

		tx.Commit()

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Order item returned and amount refunded",
		})
		return
	}

	orders.Status = models.OrderStatusReturned
	orders.PaymentStatus = models.PaymentStatusRefund
	if err := tx.Model(&orders).Updates(map[string]interface{}{
		"status":         orders.Status,
		"payment_status": orders.PaymentStatus,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update order status",
		})
		return
	}

	var orderItems []models.OrderItem
	if err := tx.Preload("Product").Where("order_id = ?", orderId).Find(&orderItems).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "order items not found for this order",
		})
		return
	}

	for _, orderItem := range orderItems {
		orderItem.Status = models.OrderStatusReturned
		if err := tx.Model(&orderItem).Update("status", orderItem.Status).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to update order item status",
			})
			return
		}

		orderItem.Product.Availability = true
		if err := tx.Model(&orderItem.Product).Update("availability", orderItem.Product.Availability).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to update product availability",
			})
			return
		}
	}

	if orders.PaymentStatus == models.PaymentStatusPaid {
		err := RefundToUser(tx, userIDStr, orderId, orders.FinalAmount, "Entire order canceled", isSeller)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to refund amount",
			})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order canceled and amount refunded",
	})
}
