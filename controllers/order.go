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
	var TotalCategoryDiscount float64
	var TotalProductOfferAmount float64
	var ProductOfferAmount float64
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

		discountedPrice = calculateFinalAmount(Product.OfferAmount, category.OfferPercentage)

		TotalAmount += Product.Price

		ProductOfferAmount = Product.Price - Product.OfferAmount
		TotalProductOfferAmount += ProductOfferAmount

		CategoryDiscount := Product.OfferAmount - discountedPrice
		TotalCategoryDiscount += CategoryDiscount

		finalAmount += Product.OfferAmount

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

	finalAmount -= TotalCategoryDiscount

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

	if PaymentMethodOption == models.COD && finalAmount > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "COD is not applicable for order",
		})
		return
	}

	deliveryCharge := 0
	if finalAmount < 500 {
		deliveryCharge = 40
		finalAmount += float64(deliveryCharge)
	}

	status := models.OrderStatusPending

	if PaymentMethodOption == models.COD {
		status = models.OrderStatusConfirmed
	}

	tx := database.DB.Begin()

	order := models.Order{
		UserID:                 userIDStr,
		TotalAmount:            RoundDecimalValue(TotalAmount),
		FinalAmount:            RoundDecimalValue(finalAmount),
		PaymentMethod:          PaymentMethodOption,
		PaymentStatus:          models.OrderStatusPending,
		OrderedAt:              time.Now(),
		CouponCode:             request.CouponCode,
		CouponDiscountAmount:   RoundDecimalValue(CouponDiscount),
		ProductOfferAmount:     RoundDecimalValue(TotalProductOfferAmount),
		DeliveryCharge:         float64(deliveryCharge),
		CategoryDiscountAmount: RoundDecimalValue(TotalCategoryDiscount),
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

	if !CartToOrderItems(userIDStr, order, CouponDiscount) {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to transfer cart items to order.",
		})
		return
	}

	if PaymentMethodOption == models.Wallet {
		orderIDStr := fmt.Sprintf("%d", order.OrderID)
		if _, err := ProcessWalletPayment(userIDStr, orderIDStr, tx); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": err.Error(),
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

func CartToOrderItems(UserID uint, Order models.Order, CouponDiscount float64) bool {
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

	tx := database.DB.Begin()

	for _, cartItem := range CartItems {
		Product := cartItem.Product

		var category models.Category
		if err := database.DB.First(&category, Product.CategoryID).Error; err != nil {
			tx.Rollback()
			return false
		}

		discountedPrice := calculateFinalAmount(Product.OfferAmount, category.OfferPercentage)
		productOffer := Product.Price - Product.OfferAmount

		categoryOffer := Product.OfferAmount - discountedPrice
		finalPrice := Product.Price - categoryOffer - productOffer

		var proportionalDiscount float64
		if CouponDiscount > 0 {
			proportionalDiscount = (Product.OfferAmount / totalCartPrice) * CouponDiscount
			finalPrice -= proportionalDiscount
		}

		finalPrice = math.Max(0, finalPrice)

		orderStatus := models.OrderStatusPending

		if Order.PaymentMethod == models.COD {
			orderStatus = models.OrderStatusConfirmed
		}

		orderItem := models.OrderItem{
			OrderID:             Order.OrderID,
			ProductID:           cartItem.ProductID,
			UserID:              UserID,
			SellerID:            Product.SellerID,
			Price:               Product.Price,
			ProductOfferAmount:  RoundDecimalValue(productOffer),
			CategoryOfferAmount: RoundDecimalValue(categoryOffer),
			OtherOffers:         RoundDecimalValue(proportionalDiscount),
			FinalAmount:         RoundDecimalValue(finalPrice),
			Status:              orderStatus,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return false
		}

		if Order.PaymentMethod == models.COD {
			Product.Availability = false
			if err := tx.Model(&Product).Where("id = ?", Product.ID).Update("availability", Product.Availability).Error; err != nil {
				tx.Rollback()
				return false
			}
		}
	}

	if Order.PaymentMethod == models.COD {
		if err := tx.Where("user_id = ?", UserID).Delete(&models.Cart{}).Error; err != nil {
			tx.Rollback()
			return false
		}
	}

	if err := tx.Commit().Error; err != nil {
		return false
	}

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
				FinalAmount: item.FinalAmount,
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
		order.PaymentStatus = models.PaymentStatusPaid
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
				Price:       RoundDecimalValue(orderItem.Price),
				FinalAmount: RoundDecimalValue(orderItem.FinalAmount),
				Image:       orderItem.Product.Image,
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
			DeliveryCharge:  order.DeliveryCharge,
			FinalAmount:     RoundDecimalValue(order.FinalAmount),
			Items:           orderItemResponses,
			Status:          order.Status,
			PaymentStatus:   order.PaymentStatus,
			PaymentMethod:   order.PaymentMethod,
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

		if orderItem.Status == models.OrderStatusCanceled {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "this item is already cancelled",
			})
			return
		}

		orders.FinalAmount -= orderItem.FinalAmount
		if err := tx.Model(&orders).Update("final_amount", orders.FinalAmount).Error; err != nil {
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
			err := RefundToUser(tx, id, orderId, orderItem.FinalAmount, "Single item canceled", isSeller)
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

		var itemCount int64
		if err := tx.Model(&models.OrderItem{}).Where("order_id = ? AND status != ?", orderId, models.OrderStatusCanceled).Count(&itemCount).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to check order items",
			})
			return
		}

		if itemCount == 0 {
			orders.Status = models.OrderStatusCanceled
			if err := tx.Model(&orders).Update("status", orders.Status).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "failed",
					"message": "failed to update order status",
				})
				return
			}
		}

		tx.Commit()

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Order item canceled and amount refunded",
		})
		return
	}

	if orders.Status == models.OrderStatusCanceled {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "this order is already cancelled",
		})
		return
	}

	if orders.PaymentStatus == models.PaymentStatusPaid {
		fmt.Println("refund is starting")
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

		if orderItem.Status != models.OrderStatusDelivered {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "item not delivered yet",
			})
			return
		}

		if orderItem.Status == models.OrderStatusReturned {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "this item is already returned",
			})
			return
		}

		if orders.PaymentStatus == models.PaymentStatusPaid {
			err := RefundToUser(tx, userIDStr, orderId, orderItem.FinalAmount, "Single item returned", isSeller)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "failed",
					"message": "failed to refund amount",
				})
				return
			}
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

		orders.FinalAmount -= orderItem.FinalAmount
		if err := tx.Model(&orders).Update("final_amount", orders.FinalAmount).Error; err != nil {
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

		var itemCount int64
		if err := tx.Model(&models.OrderItem{}).Where("order_id = ? AND status != ?", orderId, models.OrderStatusReturned).Count(&itemCount).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to check order items",
			})
			return
		}

		if itemCount == 0 {
			orders.Status = models.OrderStatusReturned
			if err := tx.Model(&orders).Update("status", orders.Status).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "failed",
					"message": "failed to update order status",
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

	if orders.Status == models.OrderStatusReturned {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "this order is already returned",
		})
		return
	}

	if orders.Status != models.OrderStatusDelivered {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "order is not delivered yet",
		})
		return
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

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order canceled and amount refunded",
	})
}
