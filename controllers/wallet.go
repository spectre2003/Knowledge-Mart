package controllers

import (
	"errors"
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"strconv"
	"time"

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

	Seller.WalletAmount += Order.FinalAmount

	sellerWallet := models.SellerWallet{
		TransactionTime: time.Now(),
		Type:            models.WalletIncoming,
		OrderID:         Order.OrderID,
		SellerID:        Order.SellerID,
		Amount:          Order.FinalAmount,
		CurrentBalance:  Seller.WalletAmount,
		Reason:          "Order Payment",
	}

	if err := database.DB.Create(&sellerWallet).Error; err != nil {
		return false
	}

	if err := database.DB.Model(&Seller).Update("wallet_amount = ?", sellerWallet.CurrentBalance).Error; err != nil {
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

	var User models.User

	if err := tx.Where("id = ?", userID).First(&User).Error; err != nil {
		return fmt.Errorf("failed to find the user: %w", err)
	}

	if err := tx.Model(&User).Where("id = ?", userID).Update("wallet_amount", wallet.CurrentBalance).Error; err != nil {
		return fmt.Errorf("failed to update user wallet balance: %w", err)
	}
	var order models.Order
	if err := tx.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return fmt.Errorf("failed to find order with ID %d: %w", orderID, err)
	}

	if order.SellerID == 0 {
		return fmt.Errorf("invalid seller ID in order ID %d", orderID)
	}

	var seller models.Seller
	if err := tx.Where("id = ?", order.SellerID).First(&seller).Error; err != nil {
		return fmt.Errorf("failed to find seller with ID %d: %w", order.SellerID, err)
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

	if err := tx.Model(&seller).Where("id = ?", seller.ID).Update("wallet_amount", sellerWalletTransaction.CurrentBalance).Error; err != nil {
		return fmt.Errorf("failed to update seller wallet amount: %w", err)
	}

	if err := tx.Create(&sellerWalletTransaction).Error; err != nil {
		return fmt.Errorf("failed to create seller wallet transaction: %w", err)
	}

	return nil
}

func ProcessWalletPayment(userID uint, orderIDStr string, tx *gorm.DB) (models.UserWallet, error) {
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

	var OrderItem []models.OrderItem

	if err := tx.Where("order_id = ?", orderIDStr).First(&OrderItem).Error; err != nil {
		return models.UserWallet{}, fmt.Errorf("failed to find the order item")
	}

	for _, orderItem := range OrderItem {
		orderItem.Status = models.OrderStatusConfirmed
		if err := tx.Model(&orderItem).Update("status", orderItem.Status).Error; err != nil {
			return models.UserWallet{}, fmt.Errorf("failed to update order item status for item ID")
		}
	}
	return newUserWallet, nil
}
