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
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

func RenderRazorpay(c *gin.Context) {
	c.HTML(http.StatusOK, "payment.html", nil)
}

func CreateOrder(c *gin.Context) {
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	var order models.Order
	// Fetch the order from the database; consider handling the error.
	if err := database.DB.Model(&models.Order{}).Where("order_id=?", 8).First(&order).Error; err != nil {
		fmt.Println("Error fetching order:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order"})
		return
	}

	// Convert TotalAmount to an integer in paise
	amount := int(order.TotalAmount * 100) // Explicitly convert to int64

	// Create the Razorpay order
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
	orderid := strconv.Itoa(o_id)
	fmt.Println(orderid)

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

	// Get the Razorpay secret key from environment variables
	secret := os.Getenv("RAZORPAY_KEY_SECRET")

	// Verify payment signature using HMAC-SHA256
	if !verifySignature(paymentInfo.OrderID, paymentInfo.PaymentID, paymentInfo.Signature, secret) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment signature"})
		return
	}

	// Save payment details to the database
	payment := models.Payment{
		OrderID:           orderid,
		WalletPaymentID:   "",
		RazorpayOrderID:   paymentInfo.OrderID,
		RazorpayPaymentID: paymentInfo.PaymentID,
		RazorpaySignature: paymentInfo.Signature,
		PaymentGateway:    "Razorpay",
		PaymentStatus:     "PAID",
	}
	fmt.Println(payment)

	// Save payment to the database
	database.DB.Model(&models.Payment{}).Create(&payment)

	// Update the order's payment status
	database.DB.Model(&models.Order{}).Where("order_id=?", orderid).Update("payment_status", payment.PaymentStatus)

	// Reset order ID
	o_id = 0

	// Payment verified successfully
	c.JSON(http.StatusOK, gin.H{"status": "Payment verified successfully"})
}

func verifySignature(orderID, paymentID, signature, secret string) bool {
	// Concatenate the Razorpay Order ID and Payment ID
	data := orderID + "|" + paymentID

	// Create a new HMAC by defining the hash type and the secret key
	h := hmac.New(sha256.New, []byte(secret))

	// Write the data to the HMAC
	h.Write([]byte(data))

	// Get the computed HMAC in hex format
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare the computed HMAC with the provided Razorpay signature
	return expectedSignature == signature
}
