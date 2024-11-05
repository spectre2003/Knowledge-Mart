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
	OfferAmount float64        `validate:"required,number" json:"offer_amount"`
	Image       pq.StringArray `validate:"required,dive,url" json:"image_url"`
}

type EditProductRequest struct {
	ProductID    uint           `validate:"required" json:"productId"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	OfferAmount  float64        `json:"offer_amount"`
	Image        pq.StringArray `json:"image_url"`
	Availability *bool          `json:"availability"`
	CategoryID   uint           `json:"categoryid"`
}

type AddCategoryRequest struct {
	Name            string `validate:"required" json:"name"`
	Description     string `validate:"required" json:"description"`
	Image           string `validate:"required" json:"image"`
	OfferPercentage uint   `json:"offer_percentage"`
}

type EditCategoryRequest struct {
	ID              uint   `validate:"required,number" json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Image           string `json:"image"`
	OfferPercentage uint   `json:"offer_percentage"`
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
	Picture     string `json:"picture"`
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

type ChangeOrderStatusRequest struct {
	OrderItemID uint `validate:"required,number" json:"orderItemId"`
	//Status      string `validate:"required" json:"status"`
}

type RatingRequest struct {
	SellerID uint    `json:"seller_id" binding:"required"`
	Rating   float64 `json:"rating" binding:"required,min=1,max=5"`
}

type PlaceOrder struct {
	AddressID     uint   `validate:"required,number" json:"address_id"`
	PaymentMethod uint   `validate:"required" json:"payment_method"`
	CouponCode    string `json:"coupon_code"`
}

type RazorpayPayment struct {
	OrderID   uint   `json:"order_id"`
	PaymentID string `json:"payment_id"`
	Signature string `json:"signature"`
}

type CouponInventoryRequest struct {
	//CouponCode            string `validate:"required" json:"coupon_code"`
	Expiry                int64   `validate:"required" json:"expiry"`
	Percentage            uint    `validate:"required" json:"percentage"`
	MaximumUsage          uint    `validate:"required" json:"maximum_usage"`
	MinimumAmount         float64 `validate:"required" json:"minimum_amount"`
	MaximumDiscountAmount float64 `json:"maximum_discount_amount"`
}

type UpdateCouponInventoryRequest struct {
	CouponCode            string  `validate:"required" json:"coupon_code"`
	Expiry                int64   `validate:"required" json:"expiry"`
	Percentage            uint    `validate:"required" json:"percentage"`
	MaximumUsage          uint    `validate:"required" json:"maximum_usage"`
	MinimumAmount         float64 `validate:"required" json:"minimum_amount"`
	MaximumDiscountAmount float64 `validate:"required" json:"maximum_discount_amount"`
}

type AddOfferRequest struct {
	ProductID   uint    `json:"product_id" binding:"required"`
	OfferAmount float64 `json:"offer_amount" binding:"required"`
}

type CreateNoteSharing struct {
	Name string `json:"name" validate:"required"`
}

type SemesterRequest struct {
	Number int `json:"number" validate:"required"`
}

type UploadNote struct {
	CourseID    uint   `json:"course_id" validate:"required"`
	SemesterID  uint   `json:"semester_id" validate:"required"`
	SubjectID   uint   `json:"subject_id" validate:"required"`
	Description string `json:"description" validate:"required"`
	FileURL     string `json:"file_url" validate:"required"`
}
