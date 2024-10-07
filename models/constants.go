package models

const (
	OrderStatusPending        = "pending"
	OrderStatusShipped        = "shipped"
	OrderStatusOutForDelivery = "outForDelivery"
	OrderStatusDelivered      = "delivered"
	OrderStatusCanceled       = "canceled"

	PaymentStatusPaid     = "Paid"
	PaymentStatusCanceled = "Canceled"

	OnlinePaymentPending   = "Pending"
	OnlinePaymentConfirmed = "Confirmed"
	OnlinePaymentFailed    = "Failed"

	Razorpay = "Razorpay"
)
