package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type ProductResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	OfferAmount float64 `json:"offer_amount"`
	//FinalAmount  float64        `json:"final_amount"`
	Image        pq.StringArray `json:"image_url"`
	Availability bool           `json:"availability"`
	SellerID     uint           `json:"sellerid"`
	CategoryID   uint           `json:"categoryid"`
	SellerRating float64        `json:"sellerRating"`
}

type ProductCategoryResponse struct {
	ID           uint           `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	OfferAmount  float64        `json:"offer_amount"`
	FinalAmount  float64        `json:"final_amount"`
	Image        pq.StringArray `json:"image_url"`
	Availability bool           `json:"availability"`
	SellerID     uint           `json:"sellerid"`
	CategoryID   uint           `json:"categoryid"`
	SellerRating float64        `json:"sellerRating"`
}

type UserResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Picture     string `json:"picture"`
	Blocked     bool   `json:"blocked"`
	IsVerified  bool   `json:"verified"`
}

type SellerResponse struct {
	ID           uint    `json:"id"`
	UserID       uint    `json:"userid"`
	User         string  `json:"user"`
	Email        string  `json:"email"`
	UserName     string  `json:"name"`
	PhoneNumber  string  `json:"phone_number"`
	Description  string  `json:"description"`
	IsVerified   bool    `json:"verified"`
	SellerRating float64 `json:"sellerRating"`
}

type CatgoryResponse struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Image           string `json:"image"`
	OfferPercentage uint   `json:"offer_percentage"`
}

type GoogleResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type AddressResponse struct {
	ID           uint   `json:"id"`
	StreetName   string `json:"streetname"`
	StreetNumber string `json:"streetNumber"`
	City         string `json:"city"`
	State        string `json:"state"`
	PinCode      string `json:"pincode"`
	PhoneNumber  string `json:"phoneNumber"`
}

type UserProfileResponse struct {
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	PhoneNumber  string  `json:"phone_number"`
	Picture      string  `json:"picture"`
	ReferralCode string  `json:"referral_code"`
	WalletAmount float64 `json:"wallet_amount"`
}

type CartResponse struct {
	ProductID    uint           `json:"productId"`
	ProductName  string         `json:"productName"`
	CategoryID   uint           `json:"categoryId"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	OfferAmount  float64        `json:"offer_amount"`
	FinalAmount  float64        `json:"final_amount"`
	Availability bool           `json:"availability"`
	Image        pq.StringArray `json:"image_url"`
	ID           uint           `json:"Id"`
	SellerRating float64        `json:"sellerRating"`
}

type GetSellerOrdersResponse struct {
	OrderID         uint                `json:"orderId"`
	UserID          uint                `json:"userId"`
	SellerID        uint                `json:"sellerId"`
	PaymentMethod   string              `json:"paymentMethod"`
	PaymentStatus   string              `json:"paymentStatus"`
	OrderStatus     string              `json:"orderStatus"`
	TotalAmount     float64             `json:"totalAmount"`
	FinalAmount     float64             `json:"finalAmount"`
	Product         []OrderItemResponse `json:"products"`
	ShippingAddress ShippingAddress     `json:"shippingAddress"`
}

type GetSellerOrderStatusResponse struct {
	OrderItemID uint    `json:"orderItemId"`
	ProductID   uint    `json:"productId"`
	SellerID    uint    `json:"sellerId"`
	SellerName  string  `json:"sellerName"`
	Status      string  `json:"status"`
	Price       float64 `json:"price"`
}

type UserOrderResponse struct {
	OrderID         uint                `json:"orderId"`
	OrderedAt       time.Time           `json:"orderedAt"`
	TotalAmount     float64             `json:"total_amount"`
	DeliveryCharge  float64             `json:"delivery_charge"`
	FinalAmount     float64             `json:"final_amount"`
	ShippingAddress ShippingAddress     `json:"shippingAddress"`
	Status          string              `json:"orderStatus"`
	PaymentStatus   string              `json:"paymentStatus"`
	Items           []OrderItemResponse `json:"items"`
	ItemCounts      gin.H               `json:"item_counts"`
}

type OrderItemResponse struct {
	OrderItemID uint           `json:"orderItemId"`
	ProductName string         `json:"productName"`
	Image       pq.StringArray `json:"image_url"`
	CategoryID  uint           `json:"categoryId"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	FinalAmount float64        `json:"finalAmount"`
	OrderStatus string         `json:"orderStatus"`
}

type SellerOverallSalesReport struct {
	StartDate     string `json:"start_date,omitempty" time_format:"2006-01-02"`
	EndDate       string `json:"end_date,omitempty" time_format:"2006-01-02"`
	Limit         string `json:"limit,omitempty"`
	PaymentStatus string `json:"payment_status"`
}

type OrderCount struct {
	TotalOrder     uint `json:"total_order"`
	TotalPending   uint `json:"total_pending"`
	TotalConfirmed uint `json:"total_confirmed"`
	TotalShipped   uint `json:"total_shipped"`
	TotalDelivered uint `json:"total_delivered"`
	TotalCancelled uint `json:"total_cancelled"`
	TotalReturned  uint `json:"total_returned"`
}

type AmountInformation struct {
	TotalAmountBeforeDeduction  float64 `json:"total_amount_before_deduction"`
	TotalCouponDeduction        float64 `json:"total_coupon_deduction"`
	TotalCategoryOfferDeduction float64 `json:"total_category_offer_deduction"`
	TotalProuctOfferDeduction   float64 `json:"total_product_offer_deduction"`
	TotalDeliveryCharges        float64 `json:"total_delivery_charge"`
	TotalAmountAfterDeduction   float64 `json:"total_amount_after_deduction"`
}
