package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

func RenderRazorpay(c *gin.Context) {
	orderID := c.Query("orderID")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}
	c.HTML(http.StatusOK, "payment.html", gin.H{
		"orderID": orderID,
	})
	fmt.Println(orderID)
}

func CreateOrder(c *gin.Context) {
	fmt.Println("order starting")
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	orderIDStr := c.Param("orderID")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}
	fmt.Println("orderid=" + orderIDStr)
	var order models.Order

	if err := database.DB.Model(&models.Order{}).Where("order_id=?", orderIDStr).First(&order).Error; err != nil {
		fmt.Println("Error fetching order:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order"})
		return
	}

	if order.PaymentMethod != models.Razorpay {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "you chose another payment method",
		})
		return
	}

	amount := int(order.FinalAmount * 100)

	razorpayOrder, err := client.Order.Create(map[string]interface{}{
		"amount":   amount,
		"currency": "INR",
		"receipt":  "order_rcptid_11",
	}, nil)

	fmt.Println("amount:", amount)
	if err != nil {
		fmt.Println("Error creating Razorpay order:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": razorpayOrder["id"],
		"amount":   amount,
		"currency": "INR",
	})
	fmt.Println(razorpayOrder)
}

func VerifyPayment(c *gin.Context) {
	fmt.Println("VerifyPayment started")
	orderIDStr := c.Param("orderID")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	var paymentInfo struct {
		PaymentID string `json:"razorpay_payment_id"`
		OrderID   string `json:"razorpay_order_id"`
		Signature string `json:"razorpay_signature"`
	}

	if err := c.BindJSON(&paymentInfo); err != nil {
		fmt.Println("binding error")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment information"})
		return
	}

	fmt.Println("Payment Info:", paymentInfo)

	// var order models.Order
	// if err := database.DB.Where("order_id = ?", orderIDStr).First(&order).Error; err != nil {
	// 	fmt.Println("Failed to retrieve order:", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order"})
	// 	return
	// }

	//couponDiscount := order.CouponDiscountAmount

	// if !CartToOrderItems(order.UserID, order, couponDiscount) {
	// 	database.DB.Delete(&order)
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"status":  "failed",
	// 		"message": "Failed to transfer cart items to order",
	// 	})
	// 	return
	// }

	payment := models.Payment{
		OrderID:           orderIDStr,
		WalletPaymentID:   "",
		RazorpayOrderID:   paymentInfo.OrderID,
		RazorpayPaymentID: paymentInfo.PaymentID,
		RazorpaySignature: paymentInfo.Signature,
		PaymentGateway:    models.Razorpay,
		PaymentStatus:     models.OnlinePaymentPending,
	}

	if err := database.DB.Model(&models.Payment{}).Create(&payment).Error; err != nil {
		fmt.Println("Failed to create payment record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to create payment record: " + err.Error(),
		})
		return
	}

	secret := os.Getenv("RAZORPAY_KEY_SECRET")
	if verifySignature(paymentInfo.OrderID, paymentInfo.PaymentID, paymentInfo.Signature, secret) {
		if err := database.DB.Model(&models.Order{}).
			Where("order_id = ?", orderIDStr).
			Updates(map[string]interface{}{
				"payment_status": models.PaymentStatusPaid,
				"status":         models.OrderStatusConfirmed,
			}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Failed to update order payment and status",
			})
			return
		}
		if err := database.DB.Model(&models.OrderItem{}).
			Where("order_id = ?", orderIDStr).
			Update("status", models.OrderStatusConfirmed).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Failed to update order item status",
			})
			return
		}

		database.DB.Model(&models.Payment{}).
			Where("order_id = ?", orderIDStr).
			Update("payment_status", models.PaymentStatusPaid)

		if !AddMoneyToSellerWallet(orderIDStr) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Failed to add money to seller wallet",
			})
			return
		}

		fmt.Println("Payment verified successfully, order confirmed")
		c.JSON(http.StatusOK, gin.H{"status": "Payment verified successfully"})
	} else {
		fmt.Println("Invalid payment signature, order marked as pending")
		c.JSON(http.StatusOK, gin.H{"status": "Payment verification failed, order marked as pending"})
	}
}

func verifySignature(orderID, paymentID, signature, secret string) bool {
	data := orderID + "|" + paymentID

	h := hmac.New(sha256.New, []byte(secret))

	h.Write([]byte(data))

	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return expectedSignature == signature
}
