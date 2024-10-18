package controllers

import (
	"errors"
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func CreateCoupen(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	var Request models.CouponInventoryRequest
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	if CheckCouponExists(Request.CouponCode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "coupon code already exists",
		})
		return
	}

	if Request.Percentage > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "coupon discount percentage should not exceed more than 50%",
		})
		return
	}

	expiryDays := int64(Request.Expiry)
	expiryTimestamp := time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour).Unix()

	if expiryTimestamp < time.Now().Unix()+12*3600 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "please change the expiry time that is more than a day",
		})
		return
	}

	Coupon := models.CouponInventory{
		CouponCode:    Request.CouponCode,
		Expiry:        expiryTimestamp,
		Percentage:    Request.Percentage,
		MaximumUsage:  Request.MaximumUsage,
		MinimumAmount: float64(Request.MinimumAmount),
	}

	if err := database.DB.Create(&Coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to create coupon",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully created coupon",
	})
}

func UpdateCoupon(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	var Request models.CouponInventoryRequest
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var existingCoupon models.CouponInventory
	err := database.DB.Where("coupon_code = ?", Request.CouponCode).First(&existingCoupon).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "coupon not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to find coupon",
		})
		return
	}

	expiryDuration := time.Duration(Request.Expiry) * 24 * time.Hour
	newExpiryTime := time.Now().Add(expiryDuration).Unix()

	if newExpiryTime <= time.Now().Add(24*time.Hour).Unix() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "please ensure expiry is more than 1 day from now",
		})
		return
	}

	existingCoupon.Expiry = newExpiryTime
	existingCoupon.Percentage = Request.Percentage
	existingCoupon.MaximumUsage = Request.MaximumUsage

	if err := database.DB.Save(&existingCoupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update coupon",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully updated coupon",
	})
}

func DeleteCoupon(c *gin.Context) {

	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	couponCode := c.Query("coupon_code")
	if couponCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "coupon code is required",
		})
		return
	}

	var coupon models.CouponInventory
	if err := database.DB.Where("coupon_code = ?", couponCode).First(&coupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "coupon not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to find coupon",
		})
		return
	}

	if err := database.DB.Delete(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to delete coupon",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "coupon successfully deleted",
	})
}

func GetAllCoupons(c *gin.Context) {
	var Coupons []models.CouponInventory

	if err := database.DB.Find(&Coupons).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to fetch coupon details",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   Coupons,
	})
}

func ApplyCouponOnCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	UserIDStr, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	CouponCode := c.Query("couponcode")

	if CouponCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "coupon code is required",
		})
		return
	}

	var CartItems []models.Cart
	if err := database.DB.Preload("Product").Where("user_id = ?", UserIDStr).Find(&CartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to fetch cart items. Please try again later.",
		})
		return
	}

	if len(CartItems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "Your cart is empty.",
		})
		return
	}

	var sum float64
	var CartResponse []models.CartResponse

	for _, item := range CartItems {
		var Product models.Product

		if err := database.DB.Preload("Seller").Where("id = ?", item.ProductID).First(&Product).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "Failed to fetch product information. Please try again later.",
			})
			return
		}

		CartResponse = append(CartResponse, models.CartResponse{
			ProductID:    item.ProductID,
			ProductName:  item.Product.Name,
			CategoryID:   item.Product.CategoryID,
			Description:  item.Product.Description,
			Price:        item.Product.Price,
			Availability: item.Product.Availability,
			Image:        item.Product.Image,
			SellerRating: Product.Seller.AverageRating,
			ID:           item.ID,
		})

		//ProductOfferAmount += float64(ProductOfferAmount) * float64()
		sum += Product.Price

	}
	var couponDiscount float64
	var finalAmount float64

	if CouponCode != "" {
		var coupon models.CouponInventory
		if err := database.DB.Where("coupon_code = ?", CouponCode).First(&coupon).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "Invalid coupon code. Please check and try again.",
			})
			return
		}

		if time.Now().Unix() > int64(coupon.Expiry) {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "The coupon has expired.",
			})
			return
		}

		if sum < coupon.MinimumAmount {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "minimum " + strconv.Itoa(int(coupon.MinimumAmount)) + " is needed for using this coupon",
			})
			return
		}

		var usage models.CouponUsage
		usageErr := database.DB.Where("user_id = ? AND coupon_code = ?", UserIDStr, CouponCode).First(&usage).Error

		if usageErr != nil && !errors.Is(usageErr, gorm.ErrRecordNotFound) {
			// Any error other than "record not found" is an issue
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "Error checking coupon usage. Please try again later.",
			})
			return
		}

		if usageErr == nil {
			// If a record exists, it means the user has already used the coupon
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "The coupon usage limit has been reached.",
			})
			return
		}

		couponDiscount = float64(sum) * (float64(coupon.Percentage) / 100.0)
		finalAmount = sum - couponDiscount
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"cart_items":      CartResponse,
			"total_amount":    fmt.Sprintf("%.2f", sum),
			"coupon_discount": fmt.Sprintf("%.2f", couponDiscount),
			//"product_offer_amount": ProductOfferAmount,
			"final_amount": fmt.Sprintf("%.2f", finalAmount),
		},
		"message": "Cart items retrieved successfully",
	})
}

func ApplyCouponToOrder(TotalAmount float64, UserID uint, CouponCode string) (bool, string, float64) {
	// Find the coupon by code
	var coupon models.CouponInventory
	if err := database.DB.Where("coupon_code = ?", CouponCode).First(&coupon).Error; err != nil {
		return false, "coupon not found", 0
	}

	// Check for coupon expiration
	if time.Now().Unix() > int64(coupon.Expiry) {
		return false, "coupon has expired", 0
	}

	// Check coupon usage by the user
	var couponUsage models.CouponUsage
	err := database.DB.Where("coupon_code = ? AND user_id = ?", CouponCode, UserID).First(&couponUsage).Error
	if err == nil && couponUsage.UsageCount >= coupon.MaximumUsage {
		return false, "coupon usage limit reached", 0
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return false, "database error", 0
	}

	// Check if order qualifies for the coupon (minimum order amount)
	if TotalAmount < coupon.MinimumAmount {
		errMsg := fmt.Sprintf("minimum of %v is needed for using this coupon", coupon.MinimumAmount)
		return false, errMsg, 0
	}

	// Calculate discount
	discountAmount := TotalAmount * float64(coupon.Percentage) / 100

	// Update or create the coupon usage record
	if err == gorm.ErrRecordNotFound {
		// Create a new coupon usage record
		couponUsage = models.CouponUsage{
			UserID:     UserID,
			CouponCode: CouponCode,
			UsageCount: 1,
		}
		if err := database.DB.Create(&couponUsage).Error; err != nil {
			return false, "failed to create coupon usage record", 0
		}
	} else {
		// Update the coupon usage count
		couponUsage.UsageCount++
		if err := database.DB.Where("user_id = ? AND coupon_code = ?", UserID, CouponCode).Save(&couponUsage).Error; err != nil {
			return false, "failed to update coupon usage record", 0
		}
	}

	return true, "coupon applied successfully", discountAmount
}

func CheckCouponExists(code string) bool {
	var Coupons []models.CouponInventory

	if err := database.DB.Find(&Coupons).Error; err != nil {
		return false
	}

	for _, c := range Coupons {
		if c.CouponCode == code {
			return true
		}
	}

	return false
}
