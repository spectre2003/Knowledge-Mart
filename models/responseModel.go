package models

import (
	"time"

	"github.com/lib/pq"
)

type ProductResponse struct {
	ID           uint           `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
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
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
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
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Picture     string `json:"picture"`
}

type CartResponse struct {
	ProductID    uint           `json:"productId"`
	ProductName  string         `json:"productName"`
	CategoryID   uint           `json:"categoryId"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	Availability bool           `json:"availability"`
	Image        pq.StringArray `json:"image_url"`
	CartID       uint           `json:"cartId"`
	SellerRating float64        `json:"sellerRating"`
}

type GetSellerOrdersResponse struct {
	OrderID         uint            `json:"orderId"`
	UserID          uint            `json:"userId"`
	SellerID        uint            `json:"sellerId"`
	PaymentMethod   string          `json:"paymentMethod"`
	PaymentStatus   string          `json:"paymentStatus"`
	OrderStatus     string          `json:"orderStatus"`
	TotalAmount     float64         `json:"totalAmount"`
	Product         []ProductArray  `json:"products"`
	ShippingAddress ShippingAddress `json:"shippingAddress"`
}

type ProductArray struct {
	ProductID   uint           `json:"productId"`
	ProductName string         `json:"productName"`
	Description string         `json:"description"`
	Image       pq.StringArray `json:"image_url"`
	Price       float64        `json:"price"`
	OrderItemID uint           `json:"orderItemId"`
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
	TotalAmount     float64             `json:"totalAmount"`
	ShippingAddress ShippingAddress     `json:"shippingAddress"`
	Status          string              `json:"orderStatus"`
	PaymentStatus   string              `json:"paymentStatus"`
	Items           []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	OrderItemID uint           `json:"orderItemId"`
	ProductName string         `json:"productName"`
	Image       pq.StringArray `json:"image_url"`
	CategoryID  uint           `json:"categoryId"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	SellerName  string         `json:"sellerName"`
	OrderStatus string         `json:"orderStatus"`
}
