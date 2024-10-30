package models

const (
	OrderStatusPending        = "pending"
	OrderStatusConfirmed      = "confirmed"
	OrderStatusShipped        = "shipped"
	OrderStatusOutForDelivery = "outForDelivery"
	OrderStatusDelivered      = "delivered"
	OrderStatusCanceled       = "canceled"
	OrderStatusReturned       = "return"

	PaymentStatusPaid     = "Paid"
	PaymentStatusCanceled = "Canceled"
	PaymentStatusRefund   = "Refund"
	PaymentStatusFailed   = "Failed"

	OnlinePaymentPending   = "Pending"
	OnlinePaymentConfirmed = "Confirmed"
	OnlinePaymentFailed    = "Failed"

	Razorpay = "RAZORPAY"
	Wallet   = "WALLET"
	COD      = "COD"

	CODStatusPending   = "COD_PENDING"
	CODStatusConfirmed = "COD_CONFIRMED"
	CODStatusFailed    = "COD_FAILED"

	WalletIncoming = "INCOMING"
	WalletOutgoing = "OUTGOING"
)
