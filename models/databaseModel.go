package models

import (
	"time"

	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model
	Id       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Email    string `gorm:"type:varchar(255);unique" validate:"required,email" json:"email"`
	Password string `gorm:"type:varchar(255)" validate:"required" json:"password"`
}

type User struct {
	gorm.Model
	Id          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(255)" validate:"required" json:"name"`
	Email       string `gorm:"type:varchar(255);unique_index" validate:"email" json:"email"`
	PhoneNumber string `gorm:"type:varchar(255);unique_index" validate:"number" json:"phone_number"`
	Picture     string `gorm:"type:text" json:"picture"`
	Password    string `gorm:"type:varchar(255)" validate:"required" json:"password"`
	Blocked     bool   `gorm:"type:bool" json:"blocked"`
	Address     string `gorm:"type:varchar(255)" json:"address"`
	OTP         uint64
	OTPExpiry   time.Time
	IsVerified  bool
}
