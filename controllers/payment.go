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

var razorpayClient *razorpay.Client

func init() {
	keyID := os.Getenv("RAZORPAY_KEY_ID")
	secretID := os.Getenv("RAZORPAY_KEY_SECRET")
	razorpayClient = razorpay.NewClient(keyID, secretID)
}

func CreateOrder(c *gin.Context) {
	var initiatePayment models.InitiatePayment
	if err := c.ShouldBindJSON(&initiatePayment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order details"})
		return
	}

	// Calculate the amount in paisa (assuming you have final amount logic)
	orderAmount := initiatePayment.Amount * 100

	orderData := map[string]interface{}{
		"amount":          orderAmount, // Amount in paise
		"currency":        "INR",
		"receipt":         initiatePayment.OrderID,
		"payment_capture": 1, // Automatic capture
	}

	// Create Razorpay order
	rzpOrder, err := razorpayClient.Order.Create(orderData, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save Razorpay order details in DB
	RazorpayOrderID := rzpOrder["id"].(string)
	payment := models.Payment{
		OrderID:         initiatePayment.OrderID,
		RazorpayOrderID: RazorpayOrderID,
		PaymentGateway:  models.Razorpay,
		PaymentStatus:   models.OnlinePaymentPending,
	}

	if err := database.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save payment details"})
		return
	}

	// Send the order response back to the frontend
	responseData := map[string]interface{}{
		"razorpay_order_id": RazorpayOrderID,
		"amount":            orderAmount,
		"currency":          "INR",
		"key":               os.Getenv("RAZORPAY_KEY_ID"),
	}
	c.JSON(http.StatusOK, responseData)
}

func verifyPayment(c *gin.Context) {
	// Define the payload structure to receive from Razorpay
	var payload struct {
		RazorpayPaymentID string `json:"razorpay_payment_id"`
		RazorpayOrderID   string `json:"razorpay_order_id"`
		RazorpaySignature string `json:"razorpay_signature"`
	}

	// Bind the incoming JSON payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// Concatenate orderID and paymentID to create the data to verify
	data := fmt.Sprintf("%s|%s", payload.RazorpayOrderID, payload.RazorpayPaymentID)

	// Retrieve the Razorpay secret key from environment variables
	secret := os.Getenv("RAZORPAY_KEY_SECRET")

	// Create HMAC-SHA256 hash for signature verification
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare the expected signature with the Razorpay signature received
	if payload.RazorpaySignature != expectedSignature {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Signature verification failed"})
		return
	}

	// Signature verification succeeded, update the payment status in the database
	if err := database.DB.Model(&models.Payment{}).
		Where("razorpay_order_id = ?", payload.RazorpayOrderID).
		Update("payment_status", models.OnlinePaymentConfirmed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	// Update the order status in the database
	if err := database.DB.Model(&models.Order{}).
		Where("order_id = ?", payload.RazorpayOrderID).
		Update("order_status", models.PaymentStatusPaid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"status": "Payment successful"})
}
