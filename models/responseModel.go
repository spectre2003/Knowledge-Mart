package models

import "github.com/lib/pq"

type ProductResponse struct {
	ID           uint           `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	Image        pq.StringArray `json:"image_url"`
	Availability bool           `json:"availability"`
}

type UserResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Picture     string `json:"picture"`
	Address     string `json:"address"`
	Blocked     bool   `json:"blocked"`
	IsVerified  bool   `json:"verified"`
}

type SellerResponse struct {
	ID          uint   `json:"id"`
	User        User   `json:"user"`
	UserName    string `json:"name"`
	Description string `json:"description"`
	IsVerified  bool   `json:"verified"`
}

type CatgoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
}
