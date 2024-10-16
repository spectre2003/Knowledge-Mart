package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"time"
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
