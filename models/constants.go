package models

const (
	OrderStatusPending        = "pending"
	OrderStatusConfirmed      = "conformed"
	OrderStatusShipped        = "shipped"
	OrderStatusOutForDelivery = "outForDelivery"
	OrderStatusDelivered      = "delivered"
	OrderStatusCanceled       = "canceled"

	PaymentStatusPaid     = "Paid"
	PaymentStatusCanceled = "Canceled"

	OnlinePaymentPending   = "Pending"
	OnlinePaymentConfirmed = "Confirmed"
	OnlinePaymentFailed    = "Failed"

	Razorpay = "RAZORPAY"
	Wallet   = "WALLET"

	CODStatusPending   = "COD_PENDING"
	CODStatusConfirmed = "COD_CONFIRMED"
	CODStatusFailed    = "COD_FAILED"
)
