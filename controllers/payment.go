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
	orderID := c.Param("orderID")
	c.HTML(http.StatusOK, "payment.html", gin.H{
		"orderID": orderID,
	})
}

func CreateOrder(c *gin.Context) {
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	orderIDStr := c.Query("order_id")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}
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

	amount := int(order.TotalAmount * 100)

	razorpayOrder, err := client.Order.Create(map[string]interface{}{
		"amount":   amount,
		"currency": "INR",
		"receipt":  "order_rcptid_11",
	}, nil)

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
	// Capture the Razorpay Payment ID and other details from the frontend
	// orderid := strconv.Itoa(o_id)
	// fmt.Println(orderid)
	orderIDStr := c.Query("order_id")
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment information"})
		return
	}

	fmt.Println(paymentInfo)

	secret := os.Getenv("RAZORPAY_KEY_SECRET")

	if !verifySignature(paymentInfo.OrderID, paymentInfo.PaymentID, paymentInfo.Signature, secret) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment signature"})
		return
	}

	payment := models.Payment{
		OrderID:           orderIDStr,
		WalletPaymentID:   "",
		RazorpayOrderID:   paymentInfo.OrderID,
		RazorpayPaymentID: paymentInfo.PaymentID,
		RazorpaySignature: paymentInfo.Signature,
		PaymentGateway:    "Razorpay",
		PaymentStatus:     "PAID",
	}
	fmt.Println(payment)

	database.DB.Model(&models.Payment{}).Create(&payment)

	database.DB.Model(&models.Order{}).Where("order_id=?", orderIDStr).Update("payment_status", payment.PaymentStatus)

	if !AddMoneyToSellerWallet(orderIDStr) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to add money to seller wallet",
		})
		return
	}
	//o_id = 0

	c.JSON(http.StatusOK, gin.H{"status": "Payment verified successfully"})
}

func verifySignature(orderID, paymentID, signature, secret string) bool {
	data := orderID + "|" + paymentID

	h := hmac.New(sha256.New, []byte(secret))

	h.Write([]byte(data))

	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return expectedSignature == signature
}
