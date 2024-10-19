package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func AddToWhishList(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized",
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
	var Request models.AddToCartRequest
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

	var Product models.Product
	if err := database.DB.Where("id = ?", Request.ProductID).First(&Product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Failed to fetch product information. Please ensure the specified product exists.",
		})
		return
	}

	if !Product.Availability {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Product is not available",
		})
		return
	}

	var existingWhishList models.WhishList

	if err := database.DB.Where("product_id = ? AND user_id = ?", Request.ProductID, UserIDStr).First(&existingWhishList).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "This product is already in your whishlist.",
		})
		return
	}

	whishlist := models.WhishList{
		ProductID: Request.ProductID,
		UserID:    UserIDStr,
	}

	if err := database.DB.Create(&whishlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to add product to whishlist. Please try again later.",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Product added to whishlist successfully",
	})
}

func ListAllWhishList(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userIDUint).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	var whishlists []models.WhishList

	if err := database.DB.Preload("Product").Where("user_id = ?", userIDUint).Find(&whishlists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve whishlist information",
		})
		return
	}

	if len(whishlists) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Your whishlist is empty.",
		})
		return
	}
	var whishlistResponse []models.CartResponse
	ItemCount := 0

	for _, whishlist := range whishlists {
		ItemCount++
		var seller models.Seller
		if err := database.DB.Where("id = ?", whishlist.Product.SellerID).Select("average_rating").First(&seller).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve seller rating",
			})
			return
		}

		whishlistResponse = append(whishlistResponse, models.CartResponse{
			ProductID:    whishlist.ProductID,
			ProductName:  whishlist.Product.Name,
			CategoryID:   whishlist.Product.CategoryID,
			Description:  whishlist.Product.Description,
			Price:        whishlist.Product.Price,
			OfferAmount:  whishlist.Product.OfferAmount,
			Availability: whishlist.Product.Availability,
			Image:        whishlist.Product.Image,
			SellerRating: seller.AverageRating,
			ID:           whishlist.ID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully get all the whishlist items",
		"data": gin.H{
			"whishList": whishlistResponse,
			"itemCount": ItemCount,
		},
	})
}

func RemoveItemFromwhishlist(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	userIDStr, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}
	whishListIDStr := c.Query("whishlistid")
	whishListID, err := strconv.Atoi(whishListIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid whishlistID",
		})
		return
	}

	var whishlist models.WhishList

	if err := database.DB.First(&whishlist, whishListID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "product is not present in the whishlist",
		})
		return
	}

	if whishlist.UserID != userIDStr {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "You are not authorized to remove this item from the whishlist.",
		})
		return
	}

	if err := database.DB.Delete(&whishlist, whishListID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "unable to remove product from the whishlist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully remove product from whishlist",
	})
}
