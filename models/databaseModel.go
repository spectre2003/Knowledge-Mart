package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Email    string `gorm:"type:varchar(255);unique" validate:"required,email" json:"email"`
	Password string `gorm:"type:varchar(255)" validate:"required" json:"password"`
}

type User struct {
	gorm.Model
	ID           uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string  `gorm:"type:varchar(255)" validate:"required" json:"name"`
	Email        string  `gorm:"type:varchar(255);unique" validate:"email" json:"email"`
	PhoneNumber  string  `gorm:"type:varchar(255);unique" validate:"number" json:"phone_number"`
	Picture      string  `gorm:"type:text" json:"picture"`
	WalletAmount float64 `gorm:"column:wallet_amount;type:double precision" json:"wallet_amount"`
	Password     string  `gorm:"type:varchar(255)" validate:"required" json:"password"`
	Blocked      bool    `gorm:"type:bool" json:"blocked"`
	ReferralCode string  `gorm:"column:referral_code" json:"referral_code"`
	OTP          uint64
	OTPExpiry    time.Time
	IsVerified   bool   `gorm:"type:bool" json:"verified"`
	LoginMethod  string `gorm:"type:varchar(50)" json:"login_method"`
}

type UserReferralHistory struct {
	UserID       uint   `gorm:"column:user_id" json:"user_id"`
	ReferralCode string `gorm:"column:referral_code" json:"referral_code"`
	ReferredBy   uint   `gorm:"column:referred_by" json:"referred_by"`
	ReferClaimed bool   `gorm:"column:refer_claimed" json:"refer_claimed"`
}

type Seller struct {
	gorm.Model
	ID            uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint    `gorm:"not null;constraint:OnDelete:CASCADE;" json:"userId"`
	User          User    `gorm:"foreignKey:UserID"`
	UserName      string  `gorm:"type:varchar(255)" validate:"required" json:"name"`
	WalletAmount  float64 `gorm:"column:wallet_amount;type:double precision" json:"wallet_amount"`
	Password      string  `gorm:"type:varchar(255)" validate:"required" json:"password"`
	Description   string  `gorm:"type:varchar(255)" validate:"required" json:"description"`
	IsVerified    bool    `gorm:"type:bool" json:"verified"`
	AverageRating float64 `gorm:"type:decimal(10,2)" json:"averageRating"`
}

type Category struct {
	gorm.Model
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string    `gorm:"type:varchar(255)" validate:"required" json:"name"`
	Description     string    `gorm:"type:varchar(255)" validate:"required" json:"description"`
	OfferPercentage uint      `gorm:"column:offer_percentage" json:"offer_percentage"`
	Image           string    `gorm:"type:varchar(255)" validate:"required" json:"image"`
	Products        []Product `gorm:"foreignKey:CategoryID"`
}

type Product struct {
	gorm.Model
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	SellerID     uint           `gorm:"not null;constraint:OnDelete:CASCADE;" json:"sellerId"`
	Seller       Seller         `gorm:"foreignKey:SellerID"`
	Name         string         `gorm:"type:varchar(255)" validate:"required" json:"name"`
	CategoryID   uint           `gorm:"constraint:OnDelete:CASCADE;" json:"categoryId"`
	Category     Category       `gorm:"foreignKey:CategoryID"`
	Description  string         `gorm:"type:varchar(255)" validate:"required" json:"description"`
	Availability bool           `gorm:"type:bool;default:true" json:"availability"`
	Price        float64        `gorm:"type:decimal(10,2);not null" validate:"required" json:"price"`
	OfferAmount  float64        `gorm:"type:decimal(10,2);not null" validate:"required" json:"offer_amount"`
	Image        pq.StringArray `gorm:"type:varchar(255)[]" validate:"required" json:"image_url"`
}

type Address struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint   `gorm:"not null;constraint:OnDelete:CASCADE;" json:"userId"`
	User         User   `gorm:"foreignKey:UserID"`
	StreetName   string `gorm:"type:varchar(255)" validate:"required" json:"street_name"`
	StreetNumber string `gorm:"type:varchar(255)" validate:"required" json:"street_number"`
	City         string `gorm:"type:varchar(255)" validate:"required" json:"city"`
	State        string `gorm:"type:varchar(255)" validate:"required" json:"state"`
	PinCode      string `gorm:"type:varchar(255)" validate:"required" json:"pincode"`
	PhoneNumber  string `gorm:"type:varchar(255);unique" validate:"number" json:"phone_number"`
}

type Cart struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint    `gorm:"not null" json:"userId"`
	User      User    `gorm:"foreignKey:UserID"`
	ProductID uint    `gorm:"not null" json:"productId"`
	Product   Product `gorm:"foreignKey:ProductID"`
}

type Order struct {
	OrderID                uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID                 uint            `gorm:"not null" json:"userId"`
	CouponCode             string          `json:"coupon_code"`
	CouponDiscountAmount   float64         `validate:"required,number" json:"coupon_discount_amount"`
	ReferralDiscountAmount float64         `validate:"required,number" json:"referral_discount_amount"`
	CategoryDiscountAmount float64         `validate:"required,number" json:"category_discount_amount"`
	TotalAmount            float64         `gorm:"type:decimal(10,2);not null" json:"totalAmount"`
	FinalAmount            float64         `validate:"required,number" json:"final_amount"`
	PaymentMethod          string          `gorm:"type:varchar(100)" validate:"required" json:"paymentMethod"`
	PaymentStatus          string          `gorm:"type:varchar(100)" validate:"required" json:"paymentStatus"`
	OrderedAt              time.Time       `gorm:"autoCreateTime" json:"orderedAt"`
	ShippingAddress        ShippingAddress `gorm:"embedded" json:"shippingAddress"`
	SellerID               uint            `gorm:"not null" json:"sellerId"`
	Status                 string          `gorm:"type:varchar(100);default:'pending'" json:"status"`
}

type ShippingAddress struct {
	StreetName   string `gorm:"type:varchar(255)" json:"street_name"`
	StreetNumber string `gorm:"type:varchar(255)" json:"street_number"`
	City         string `gorm:"type:varchar(255)" json:"city"`
	State        string `gorm:"type:varchar(255)" json:"state"`
	PinCode      string `gorm:"type:varchar(20)" json:"pincode"`
	PhoneNumber  string `gorm:"type:varchar(20)" json:"phonenumber"`
}

type OrderItem struct {
	OrderItemID        uint    `gorm:"primaryKey;autoIncrement" json:"orderItemId"`
	OrderID            uint    `gorm:"not null" json:"orderId"`
	Order              Order   `gorm:"foreignKey:OrderID"`
	UserID             uint    `gorm:"not null" json:"userId"`
	User               User    `gorm:"foreignKey:UserID"`
	ProductID          uint    `gorm:"not null" json:"productId"`
	Product            Product `gorm:"foreignKey:ProductID"`
	SellerID           uint    `gorm:"not null" json:"sellerId"`
	Seller             Seller  `gorm:"foreignKey:SellerID"`
	Price              float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	ProductOfferAmount float64 `json:"product_offer_amount" csv:"ProductOfferAmount"`
	Status             string  `gorm:"type:varchar(100);default:'pending'" json:"status"`
}

type SellerRating struct {
	ID       uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID   uint    `gorm:"not null" json:"userId"`
	SellerID uint    `gorm:"not null" json:"sellerId"`
	Rating   float64 `gorm:"not null" json:"rating"`
}

type WhishList struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint    `gorm:"not null" json:"userId"`
	User      User    `gorm:"foreignKey:UserID"`
	ProductID uint    `gorm:"not null" json:"productId"`
	Product   Product `gorm:"foreignKey:ProductID"`
}

type Payment struct {
	ID                uint   `gorm:"primaryKey"`
	OrderID           string `gorm:"not null"`
	WalletPaymentID   string `json:"wallet_payment_id" gorm:"column:wallet_payment_id"`
	RazorpayOrderID   string `gorm:"not null"`
	RazorpayPaymentID string `gorm:"default:null"`
	RazorpaySignature string `gorm:"default:null"`
	PaymentGateway    string `gorm:"default:'Razorpay'"`
	PaymentStatus     string `gorm:"not null"`
	AmountPaid        float64
}

type UserWallet struct {
	TransactionTime time.Time `gorm:"autoCreateTime" json:"transaction_time"`
	WalletPaymentID string    `gorm:"column:wallet_payment_id" json:"wallet_payment_id"`
	UserID          uint      `gorm:"column:user_id" json:"user_id"`
	Type            string    `gorm:"column:type" json:"type"` //incoming //outgoing
	OrderID         string    `gorm:"column:order_id" json:"order_id"`
	Amount          float64   `gorm:"column:amount" json:"amount"`
	CurrentBalance  float64   `gorm:"column:current_balance" json:"current_balance"`
	Reason          string    `gorm:"column:reason" json:"reason"`
}

type SellerWallet struct {
	TransactionTime time.Time `gorm:"autoCreateTime" json:"transaction_time"`
	Type            string    `gorm:"column:type" json:"type"` //incoming //outgoing
	OrderID         uint      `gorm:"column:order_id" json:"order_id"`
	SellerID        uint      `gorm:"column:seller_id" json:"seller_id"`
	Amount          float64   `gorm:"column:amount" json:"amount"`
	CurrentBalance  float64   `gorm:"column:current_balance" json:"current_balance"`
	Reason          string    `gorm:"column:reason" json:"reason"`
}

type CouponInventory struct {
	CouponCode    string  `validate:"required" json:"coupon_code" gorm:"primary_key"`
	Expiry        int64   `validate:"required" json:"expiry"`
	Percentage    uint    `validate:"required" json:"percentage"`
	MaximumUsage  uint    `validate:"required" json:"maximum_usage"`
	MinimumAmount float64 `validate:"required" json:"minimum_amount"`
}

type CouponUsage struct {
	gorm.Model
	UserID     uint   `json:"user_id"`
	CouponCode string `json:"coupon_code"`
	UsageCount uint   `json:"usage_count"`
}
