package controllers

import (
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddMoneyToSellerWallet(OrderID string) bool {
	var Order models.Order
	if err := database.DB.Where("order_id = ?", OrderID).First(&Order).Error; err != nil {
		return false
	}
	var Seller models.Seller
	if err := database.DB.Where("id = ?", Order.SellerID).First(&Seller).Error; err != nil {
		return false
	}

	Seller.WalletAmount += Order.TotalAmount

	sellerWallet := models.SellerWallet{
		TransactionTime: time.Now(),
		Type:            models.WalletIncoming,
		OrderID:         Order.OrderID,
		SellerID:        Order.SellerID,
		Amount:          Order.TotalAmount,
		CurrentBalance:  Seller.WalletAmount,
		Reason:          "Order Payment",
	}

	if err := database.DB.Create(&sellerWallet).Error; err != nil {
		return false
	}

	if err := database.DB.Save(&Seller).Error; err != nil {
		return false
	}

	return true
}

func RefundToUser(tx *gorm.DB, userID uint, orderIDStr string, amount float64, reason string) error {

	orderIDUint, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid orderID format: %w", err)
	}
	orderID := uint(orderIDUint)

	var wallet models.UserWallet

	if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return err
	}
	newBalance := wallet.CurrentBalance + amount

	walletTransaction := models.UserWallet{
		UserID:          userID,
		WalletPaymentID: fmt.Sprintf("WALLET_%d", time.Now().Unix()),
		Type:            "incoming",
		OrderID:         orderIDStr,
		Amount:          amount,
		CurrentBalance:  newBalance,
		Reason:          reason,
		TransactionTime: time.Now(),
	}

	if err := tx.Create(&walletTransaction).Error; err != nil {
		return err
	}

	if err := tx.Model(&wallet).Update("current_balance", newBalance).Error; err != nil {
		return err
	}

	var order models.Order
	if err := tx.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return err
	}

	var seller models.Seller
	if err := tx.Where("id = ?", order.SellerID).First(&seller).Error; err != nil {
		return err
	}

	seller.WalletAmount -= amount

	sellerWalletTransaction := models.SellerWallet{
		TransactionTime: time.Now(),
		Type:            "outgoing",
		OrderID:         orderID,
		SellerID:        seller.ID,
		Amount:          amount,
		CurrentBalance:  seller.WalletAmount,
		Reason:          "Refund for order cancellation",
	}

	if err := tx.Model(&seller).Update("wallet_amount", seller.WalletAmount).Error; err != nil {
		return err
	}

	if err := tx.Create(&sellerWalletTransaction).Error; err != nil {
		return err
	}

	return nil
}

func WalletPayment(c *gin.Context) {
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

	var User models.User
	if err := database.DB.Where("id = ?", userIDStr).First(&User).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to find the user",
		})
		return
	}

	orderIDStr := c.Query("orderid")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	var Order models.Order
	if err := database.DB.Where("order_id = ? AND user_id = ?", orderIDStr, userIDStr).First(&Order).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to find the order for this user",
		})
		return
	}

	if Order.TotalAmount > User.WalletAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "not enough money to pay for this order",
		})
		return
	}

	if Order.PaymentMethod != models.Wallet {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "you chose another payment method",
		})
		return
	}

	newBalance := User.WalletAmount - Order.TotalAmount
	User.WalletAmount = newBalance

	newUserWallet := models.UserWallet{
		UserID:          userIDStr,
		WalletPaymentID: fmt.Sprintf("WALLET_%d", time.Now().Unix()),
		Type:            "incoming",
		OrderID:         orderIDStr,
		Amount:          Order.TotalAmount,
		CurrentBalance:  newBalance,
		Reason:          "Order using wallet",
		TransactionTime: time.Now(),
	}

	if err := database.DB.Create(&newUserWallet).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "failed to create new transaction record for user",
		})
		return
	}

	if err := database.DB.Model(&User).Update("wallet_amount", newBalance).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "failed to update the wallet amount of the user",
		})
		return
	}

	var seller models.Seller
	if err := database.DB.Where("id = ?", Order.SellerID).First(&seller).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "failed to find the seller",
		})
		return
	}

	newSellerBalance := seller.WalletAmount + Order.TotalAmount

	newSellerWallet := models.SellerWallet{
		TransactionTime: time.Now(),
		Type:            "incoming",
		OrderID:         Order.OrderID,
		SellerID:        seller.ID,
		Amount:          Order.TotalAmount,
		CurrentBalance:  newSellerBalance,
		Reason:          "order payment credited",
	}

	if err := database.DB.Create(&newSellerWallet).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "failed to create new transaction record for seller",
		})
		return
	}

	if err := database.DB.Model(&seller).Update("wallet_amount", newSellerBalance).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "failed to update the wallet amount of the seller",
		})
		return
	}

	payment := models.Payment{
		OrderID:         orderIDStr,
		WalletPaymentID: newUserWallet.WalletPaymentID,
		PaymentGateway:  "wallet",
		PaymentStatus:   "PAID",
	}

	if err := database.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "failed to create new payment record",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "wallet payment done successfully",
		"data": gin.H{
			"order_id": newUserWallet.WalletPaymentID,
			"amount":   Order.TotalAmount,
		},
	})
}
