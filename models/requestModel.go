package models

import (
	"github.com/lib/pq"
)

type EmailSignupRequest struct {
	Name            string `validate:"required" json:"name"`
	Email           string `validate:"required,email" json:"email"`
	PhoneNumber     string `validate:"required,number,len=10,numeric" json:"phone_number"`
	Password        string `validate:"required" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}

type EmailLoginRequest struct {
	Email    string `form:"email" validate:"required,email" json:"email"`
	Password string `form:"password" validate:"required" json:"password"`
}

type SellerRegisterRequest struct {
	UserName    string `validate:"required" json:"name"`
	Password    string `validate:"required" json:"password"`
	Description string `validate:"required" json:"description"`
}

type SellerLoginRequest struct {
	UserID   uint   `json:"userid"`
	UserName string `form:"" username:"required" json:"username"`
	Password string `form:"password" validate:"required" json:"password"`
}

type AddProductRequest struct {
	CategoryID  uint           `validate:"required,number" json:"categoryId"`
	Name        string         `validate:"required" json:"name"`
	Description string         `validate:"required" json:"description"`
	Price       float64        `validate:"required,number" json:"price"`
	Image       pq.StringArray `validate:"required,dive,url" json:"image_url"`
}

type EditProductRequest struct {
	ProductID    uint           `validate:"required" json:"productId"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	Image        pq.StringArray `json:"image_url"`
	Availability *bool          `json:"availability"`
	CategoryID   uint           `json:"categoryid"`
}

type AddCategoryRequest struct {
	Name        string `validate:"required" json:"name"`
	Description string `validate:"required" json:"description"`
	Image       string `validate:"required" json:"image"`
}

type EditCategoryRequest struct {
	ID          uint   `validate:"required,number" json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

type AddAddresRequest struct {
	StreetName   string `validate:"required" json:"street_name"`
	StreetNumber string `validate:"required" json:"street_number"`
	City         string `validate:"required" json:"city"`
	State        string `validate:"required" json:"state"`
	Pincode      string `validate:"required" json:"pincode"`
	PhoneNumber  string `validate:"required" json:"phone_number"`
}

type EditAddresRequest struct {
	ID           uint   `validate:"required,number" json:"id"`
	StreetName   string `json:"street_name"`
	StreetNumber string `json:"street_number"`
	City         string `json:"city"`
	State        string `json:"state"`
	Pincode      string `json:"pincode"`
	PhoneNumber  string `json:"phone_number"`
}

type EditUserProfileRequest struct {
	Name        string `json:"name"`
	Email       string `validate:"omitempty,email" json:"email"`
	PhoneNumber string `validate:"omitempty,number,len=10,numeric" json:"phone_number"`
	//Picture     string `json:"picture"`
}

type EditPasswordRequest struct {
	CurrentPassword string `validate:"required" json:"currentpassword"`
	NewPassword     string `validate:"required" json:"newpassword"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}

type AddToCartRequest struct {
	ProductID uint `validate:"required,number" json:"productId"`
}

type EditSellerProfileRequest struct {
	UserName    string `json:"username"`
	Description string `json:"description"`
}

type PlaceOrder struct {
	//UserID        uint   `json:"user_id"`
	AddressID     uint   `validate:"required,number" json:"address_id"`
	PaymentMethod string `validate:"required" json:"payment_method"`
}

type ChangeOrderStatusRequest struct {
	OrderItemID uint `validate:"required,number" json:"orderItemId"`
	//Status      string `validate:"required" json:"status"`
}
