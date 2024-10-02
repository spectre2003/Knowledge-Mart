package models

const (
	OrderStatusPending        = "Pending"
	OrderStatusShipped        = "Shipped"
	OrderStatusOutForDelivery = "OutForDelivery"
	OrderStatusDelivered      = "Delivered"
	OrderStatusCanceled       = "Canceled"
	//OrderStatusReturned  = "Returned"
)

var StatusOptions = []string{
	OrderStatusPending,
	OrderStatusShipped,
	OrderStatusDelivered,
	OrderStatusCanceled,
	OrderStatusOutForDelivery,
	// OrderStatusReturned,
}
