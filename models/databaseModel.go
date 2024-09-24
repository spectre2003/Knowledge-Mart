package models

import (
	"time"

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
	Name        string `gorm:"type:varchar(255)" validate:"required" json:"name"`
	Description string `gorm:"type:varchar(255)" validate:"required" json:"description"`
	IsVerified  bool   `gorm:"type:bool" json:"verified"`
}
