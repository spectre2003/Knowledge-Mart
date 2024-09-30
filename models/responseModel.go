package models

import "github.com/lib/pq"

type ProductResponse struct {
	ID           uint           `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	Image        pq.StringArray `json:"image_url"`
	Availability bool           `json:"availability"`
	SellerID     uint           `json:"sellerid"`
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
	ID          uint   `json:"id"`
	UserID      uint   `json:"userid"`
	User        string `json:"user"`
	Email       string `json:"email"`
	UserName    string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Description string `json:"description"`
	IsVerified  bool   `json:"verified"`
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
	Username     string `json:"name"`
	StreetName   string `json:"streetname"`
	StreetNumber string `json:"streetNumber"`
	City         string `json:"city"`
	State        string `json:"state"`
	PinCode      string `json:"pincode"`
	Phone        string `json:"phoneNumber"`
}

type UserProfileResponse struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Picture     string `json:"picture"`
}
