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
	Address     string `gorm:"type:varchar(255)" json:"address"`
	OTP         uint64
	OTPExpiry   time.Time
	IsVerified  bool `gorm:"type:bool" json:"verified"`
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
