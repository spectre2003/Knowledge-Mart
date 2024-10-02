package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"` // Changed to uppercase ID for consistency
	Email    string `gorm:"type:varchar(255);unique" validate:"required,email" json:"email"`
	Password string `gorm:"type:varchar(255)" validate:"required" json:"password"`
}

type User struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(255)" validate:"required" json:"name"`
	Email       string `gorm:"type:varchar(255);unique" validate:"email" json:"email"`
	PhoneNumber string `gorm:"type:varchar(255);unique" validate:"number" json:"phone_number"`
	Picture     string `gorm:"type:text" json:"picture"`
	Password    string `gorm:"type:varchar(255)" validate:"required" json:"password"`
	Blocked     bool   `gorm:"type:bool" json:"blocked"`
	OTP         uint64
	OTPExpiry   time.Time
	IsVerified  bool   `gorm:"type:bool" json:"verified"`
	LoginMethod string `gorm:"type:varchar(50)" json:"login_method"`
}

type Seller struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint   `gorm:"not null;constraint:OnDelete:CASCADE;" json:"userId"` // Foreign key for User
	User        User   `gorm:"foreignKey:UserID"`                                   // Association to User
	UserName    string `gorm:"type:varchar(255)" validate:"required" json:"name"`
	Password    string `gorm:"type:varchar(255)" validate:"required" json:"password"`
	Description string `gorm:"type:varchar(255)" validate:"required" json:"description"`
	IsVerified  bool   `gorm:"type:bool" json:"verified"`
}

type Category struct {
	gorm.Model
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(255)" validate:"required" json:"name"`
	Description string    `gorm:"type:varchar(255)" validate:"required" json:"description"`
	Image       string    `gorm:"type:varchar(255)" validate:"required" json:"image"`
	Products    []Product `gorm:"foreignKey:CategoryID"`
}

type Product struct {
	gorm.Model
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	SellerID     uint           `gorm:"not null;constraint:OnDelete:CASCADE;" json:"sellerId"`
	Seller       Seller         `gorm:"foreignKey:SellerID"`
	Name         string         `gorm:"type:varchar(255)" validate:"required" json:"name"`
	CategoryID   uint           `gorm:"not null;constraint:OnDelete:CASCADE;" json:"categoryId"`
	Category     Category       `gorm:"foreignKey:CategoryID"`
	Description  string         `gorm:"type:varchar(255)" validate:"required" json:"description"`
	Availability bool           `gorm:"type:bool;default:true" json:"availability"`
	Price        float64        `gorm:"type:decimal(10,2);not null" validate:"required" json:"price"`
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
}

type Cart struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint    `gorm:"not null" json:"userId"`
	User      User    `gorm:"foreignKey:UserID"`
	ProductID uint    `gorm:"not null" json:"productId"`
	Product   Product `gorm:"foreignKey:ProductID"`
}

type Order struct {
	gorm.Model
	OrderID         string          `gorm:"unique" validate:"required" json:"orderId"`
	UserID          uint            `gorm:"not null" json:"userId"`
	TotalAmount     float64         `gorm:"type:decimal(10,2);not null" json:"totalAmount"`
	PaymentMethod   string          `gorm:"type:varchar(100)" validate:"required" json:"paymentMethod"`
	PaymentStatus   string          `gorm:"type:varchar(100)" validate:"required" json:"paymentStatus"`
	OrderedAt       time.Time       `gorm:"autoCreateTime" json:"orderedAt"`
	ShippingAddress ShippingAddress `gorm:"type:json" json:"shipping_address"`
}

type ShippingAddress struct {
	StreetName   string `json:"street_name"`
	StreetNumber string `json:"street_number"`
	City         string `json:"city"`
	State        string `json:"state"`
	PinCode      string `json:"pincode"`
}

type OrderItem struct {
	gorm.Model
	OrderID   string  `gorm:"not null" json:"orderId"`
	Order     Order   `gorm:"foreignKey:OrderID"`
	UserID    uint    `gorm:"not null" json:"userId"`
	User      User    `gorm:"foreignKey:UserID"`
	ProductID uint    `gorm:"not null" json:"productId"`
	Product   Product `gorm:"foreignKey:ProductID"`
	SellerID  uint    `gorm:"not null" json:"sellerId"` // Foreign key to Seller
	Seller    Seller  `gorm:"foreignKey:SellerID"`
	Price     float64 `gorm:"type:decimal(10,2);not null" json:"price"`
}
