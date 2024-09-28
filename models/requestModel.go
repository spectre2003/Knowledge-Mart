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
	ProductID    uint           `json:"productId"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	Image        pq.StringArray `json:"image_url"`
	Availability bool           `json:"availability"`
}

type AddCategoryRequest struct {
	Name        string `validate:"required" json:"name"`
	Description string `validate:"required" json:"description"`
	Image       string `validate:"required" json:"image"`
}

type EditCategoryRequest struct {
	ID          uint   `validate:"required,number" json:"id"`
	Name        string `validate:"required" json:"name"`
	Description string `validate:"required" json:"description"`
	Image       string `validate:"required" json:"image"`
}

type AddAddresRequest struct {
	StreetName   string `validate:"required" json:"street_name"`
	StreetNumber string `validate:"required" json:"street_number"`
	City         string `validate:"required" json:"city"`
	State        string `validate:"required" json:"state"`
	Pincode      string `validate:"required" json:"pincode"`
}

type EditAddresRequest struct {
	ID           uint   `validate:"required,number" json:"id"`
	StreetName   string `validate:"required" json:"street_name"`
	StreetNumber string `validate:"required" json:"street_number"`
	City         string `validate:"required" json:"city"`
	State        string `validate:"required" json:"state"`
	Pincode      string `validate:"required" json:"pincode"`
}

type EditUserProfileResponse struct {
	ID          uint   `validate:"required,number" json:"id"`
	Name        string `validate:"required" json:"name"`
	Email       string `validate:"required,email" json:"email"`
	PhoneNumber string `validate:"required,number,len=10,numeric" json:"phone_number"`
	//Picture     string `json:"picture"`
}
