package controllers

import (
	"errors"
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
	var order models.Order
	if err := database.DB.Where("order_id = ?", OrderID).First(&order).Error; err != nil {
		fmt.Println("Error fetching order:", err)
		return false
	}

	var seller models.Seller
	if err := database.DB.Where("id = ?", order.SellerID).First(&seller).Error; err != nil {
		fmt.Println("Error fetching seller:", err)
		return false
	}

	finalAmount := RoundDecimalValue(order.FinalAmount)
	fmt.Println("Final Amount (Rounded):", finalAmount)

	if finalAmount < 0 {
		fmt.Println("Error: Negative final amount:", finalAmount)
		return false
	}

	seller.WalletAmount += finalAmount
	fmt.Println("Updated Wallet Amount:", seller.WalletAmount)

	sellerWallet := models.SellerWallet{
		TransactionTime: time.Now(),
		Type:            models.WalletIncoming,
		OrderID:         order.OrderID,
		SellerID:        order.SellerID,
		Amount:          finalAmount,
		CurrentBalance:  RoundDecimalValue(seller.WalletAmount),
		Reason:          "Order Payment",
	}

	if err := database.DB.Create(&sellerWallet).Error; err != nil {
		fmt.Println("Error creating seller wallet record:", err)
		return false
	}

	if err := database.DB.Model(&seller).Update("wallet_amount", sellerWallet.CurrentBalance).Error; err != nil {
		fmt.Println("Error updating seller's wallet balance:", err)
		return false
	}

	fmt.Println("Seller wallet updated successfully")
	return true
}

func RefundToUser(tx *gorm.DB, userID uint, orderIDStr string, amount float64, reason string, isSeller bool) error {
	orderIDUint, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid orderID format: %w", err)
	}
	orderID := uint(orderIDUint)

	var order models.Order
	if err := tx.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return fmt.Errorf("failed to find order with ID %d: %w", orderID, err)
	}

	if !isSeller {
		var wallet models.UserWallet

		err = tx.Where("user_id = ?", userID).First(&wallet).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				wallet = models.UserWallet{
					UserID:         userID,
					CurrentBalance: amount,
				}
				if err := tx.Create(&wallet).Error; err != nil {
					return fmt.Errorf("failed to create new wallet: %w", err)
				}
			} else {
				return fmt.Errorf("failed to retrieve wallet: %w", err)
			}
		} else {
			wallet.CurrentBalance += amount
		}

		walletTransaction := models.UserWallet{
			UserID:          userID,
			WalletPaymentID: fmt.Sprintf("WALLET_%d", time.Now().Unix()),
			Type:            "incoming",
			OrderID:         orderIDStr,
			Amount:          amount,
			CurrentBalance:  wallet.CurrentBalance,
			Reason:          reason,
			TransactionTime: time.Now(),
		}

		if err := tx.Create(&walletTransaction).Error; err != nil {
			return fmt.Errorf("failed to create wallet transaction: %w", err)
		}

		if err := tx.Model(&wallet).Where("user_id = ?", userID).Update("current_balance", wallet.CurrentBalance).Error; err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}

		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("wallet_amount", wallet.CurrentBalance).Error; err != nil {
			return fmt.Errorf("failed to update user wallet balance: %w", err)
		}

		var seller models.Seller
		if err := tx.Where("id = ?", order.SellerID).First(&seller).Error; err != nil {
			return fmt.Errorf("failed to find seller with ID %d: %w", order.SellerID, err)
		}

		seller.WalletAmount -= amount
		if err := tx.Model(&seller).Update("wallet_amount", seller.WalletAmount).Error; err != nil {
			return fmt.Errorf("failed to update seller wallet amount: %w", err)
		}

		sellerWalletTransaction := models.SellerWallet{
			TransactionTime: time.Now(),
			Type:            "outgoing",
			OrderID:         order.OrderID,
			SellerID:        seller.ID,
			Amount:          amount,
			CurrentBalance:  seller.WalletAmount,
			Reason:          "Refund due to user-initiated return/cancellation",
		}

		if err := tx.Create(&sellerWalletTransaction).Error; err != nil {
			return fmt.Errorf("failed to create seller wallet transaction: %w", err)
		}

		return nil
	}

	var seller models.Seller
	if err := tx.Where("id = ?", order.SellerID).First(&seller).Error; err != nil {
		return fmt.Errorf("failed to find seller with ID %d: %w", order.SellerID, err)
	}

	seller.WalletAmount -= amount
	if err := tx.Model(&seller).Update("wallet_amount", seller.WalletAmount).Error; err != nil {
		return fmt.Errorf("failed to update seller wallet amount: %w", err)
	}

	sellerWalletTransaction := models.SellerWallet{
		TransactionTime: time.Now(),
		Type:            "outgoing",
		OrderID:         order.OrderID,
		SellerID:        seller.ID,
		Amount:          amount,
		CurrentBalance:  seller.WalletAmount,
		Reason:          "Refund for order cancellation initiated by seller",
	}

	if err := tx.Create(&sellerWalletTransaction).Error; err != nil {
		return fmt.Errorf("failed to create seller wallet transaction: %w", err)
	}

	var wallet models.UserWallet
	err = tx.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = models.UserWallet{
				UserID:         userID,
				CurrentBalance: amount,
			}
			if err := tx.Create(&wallet).Error; err != nil {
				return fmt.Errorf("failed to create new wallet: %w", err)
			}
		} else {
			return fmt.Errorf("failed to retrieve wallet: %w", err)
		}
	} else {
		wallet.CurrentBalance += amount
	}

	walletTransaction := models.UserWallet{
		UserID:          userID,
		WalletPaymentID: fmt.Sprintf("WALLET_%d", time.Now().Unix()),
		Type:            "incoming",
		OrderID:         orderIDStr,
		Amount:          amount,
		CurrentBalance:  wallet.CurrentBalance,
		Reason:          reason,
		TransactionTime: time.Now(),
	}

	if err := tx.Create(&walletTransaction).Error; err != nil {
		return fmt.Errorf("failed to create wallet transaction: %w", err)
	}

	if err := tx.Model(&wallet).Where("user_id = ?", userID).Update("current_balance", wallet.CurrentBalance).Error; err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	fmt.Println("wallet amount")
	fmt.Println(wallet.CurrentBalance)

	if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("wallet_amount", wallet.CurrentBalance).Error; err != nil {
		return fmt.Errorf("failed to update user wallet balance: %w", err)
	}

	return nil
}

func ProcessWalletPayment(userID uint, orderIDStr string, couponDiscount float64, referralDiscount float64, tx *gorm.DB) (models.UserWallet, error) {
	var User models.User
	if err := tx.Where("id = ?", userID).First(&User).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to find the user")
	}

	var Order models.Order
	if err := tx.Where("order_id = ? AND user_id = ?", orderIDStr, userID).First(&Order).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to find the order for this user")
	}

	if Order.FinalAmount > User.WalletAmount {
		return models.UserWallet{}, fmt.Errorf("not enough wallet balance to pay for this order")
	}

	if Order.PaymentMethod != models.Wallet {
		return models.UserWallet{}, fmt.Errorf("incorrect payment method")
	}

	newBalance := User.WalletAmount - Order.FinalAmount
	if err := tx.Model(&User).Update("wallet_amount", newBalance).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to update user wallet balance")
	}

	newUserWallet := models.UserWallet{
		UserID:          userID,
		WalletPaymentID: fmt.Sprintf("WALLET_%d", time.Now().Unix()),
		Type:            "outgoing",
		OrderID:         orderIDStr,
		Amount:          Order.FinalAmount,
		CurrentBalance:  newBalance,
		Reason:          "Order payment using wallet",
		TransactionTime: time.Now(),
	}

	if err := tx.Create(&newUserWallet).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to create wallet transaction record")
	}

	var seller models.Seller
	if err := tx.Where("id = ?", Order.SellerID).First(&seller).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to find the seller")
	}

	newSellerBalance := seller.WalletAmount + Order.FinalAmount
	if err := tx.Model(&seller).Update("wallet_amount", newSellerBalance).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to update seller wallet balance")
	}

	newSellerWallet := models.SellerWallet{
		TransactionTime: time.Now(),
		Type:            "incoming",
		OrderID:         Order.OrderID,
		SellerID:        seller.ID,
		Amount:          Order.FinalAmount,
		CurrentBalance:  newSellerBalance,
		Reason:          "order payment credited",
	}

	if err := tx.Create(&newSellerWallet).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to create seller transaction record")
	}

	payment := models.Payment{
		OrderID:         orderIDStr,
		WalletPaymentID: newUserWallet.WalletPaymentID,
		PaymentGateway:  "wallet",
		PaymentStatus:   "PAID",
	}

	if err := tx.Create(&payment).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to create payment record")
	}

	Order.PaymentStatus = models.PaymentStatusPaid
	Order.Status = models.OrderStatusConfirmed

	if err := tx.Model(&Order).Updates(map[string]interface{}{
		"status":         Order.Status,
		"payment_status": Order.PaymentStatus,
	}).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to update payment and order status")
	}

	if !CartToOrderItems(userID, Order, couponDiscount, referralDiscount) {
		tx.Rollback()
		return models.UserWallet{}, fmt.Errorf("failed to transfer cart items to order")
	}

	var orderItems []models.OrderItem
	if err := tx.Where("order_id = ?", orderIDStr).Find(&orderItems).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to find the order items")
	}

	for _, orderItem := range orderItems {
		orderItem.Status = models.OrderStatusConfirmed
		if err := tx.Model(&orderItem).Update("status", orderItem.Status).Error; err != nil {
			return models.UserWallet{}, fmt.Errorf("failed to update order item status for item ID")
		}
	}

	return newUserWallet, nil
}

func GetUserWalletHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}
	var walletHistory []models.UserWallet
	if err := database.DB.Where("user_id = ?", uint(userIDUint)).Order("transaction_time desc").Find(&walletHistory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "failed",
				"error":  "No wallet history found for the user",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to fetch wallet history",
			"status": "failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"wallet_history": walletHistory,
	})
}

func GetSellerWalletHistory(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized",
		})
		return
	}

	sellerIDUint, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}
	var walletHistory []models.SellerWallet
	if err := database.DB.Where("seller_id = ?", uint(sellerIDUint)).Order("transaction_time desc").Find(&walletHistory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "failed",
				"error":  "No wallet history found for the seller",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to fetch wallet history",
			"status": "failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"wallet_history": walletHistory,
	})
}
