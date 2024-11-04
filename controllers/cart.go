package controllers

import (
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func AddToCart(c *gin.Context) {
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

	var seller models.Seller
	if err := database.DB.Where("id = ?", Product.SellerID).First(&seller).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Failed to fetch seller information.",
		})
		return
	}

	if UserIDStr == seller.UserID {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "you cant buy your own product",
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

	var existingCartItem models.Cart
	if err := database.DB.Where("product_id = ? AND user_id = ?", Request.ProductID, UserIDStr).First(&existingCartItem).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "This product is already in your cart.",
		})
		return
	}

	var userCartItems []models.Cart
	if err := database.DB.Where("user_id = ?", UserIDStr).Find(&userCartItems).Error; err == nil {
		for _, cartItem := range userCartItems {
			var cartProduct models.Product
			if err := database.DB.Where("id = ?", cartItem.ProductID).First(&cartProduct).Error; err == nil {
				if cartProduct.SellerID != Product.SellerID {
					c.JSON(http.StatusBadRequest, gin.H{
						"status":  "failed",
						"message": "You can only add products from one seller at a time. Please complete or clear your current cart before adding products from a different seller.",
					})
					return
				}
			}
		}
	}

	cart := models.Cart{
		ProductID: Request.ProductID,
		UserID:    UserIDStr,
	}

	if err := database.DB.Create(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to add product to cart. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Product added to cart successfully",
	})
}

func ListAllCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized",
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

	var Carts []models.Cart
	if err := database.DB.Preload("Product").Where("user_id = ?", userIDUint).Find(&Carts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve cart information",
		})
		return
	}

	if len(Carts) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Your cart is empty.",
		})
		return
	}

	var CartResponse []models.CartResponse
	var TotalAmount float64
	ItemCount := 0

	for _, cart := range Carts {
		var category models.Category
		if err := database.DB.Where("id = ?", cart.Product.CategoryID).Select("offer_percentage").First(&category).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve category offer",
			})
			return
		}

		finalAmount := calculateFinalAmount(cart.Product.OfferAmount, category.OfferPercentage)
		TotalAmount += finalAmount
		ItemCount++

		var seller models.Seller
		if err := database.DB.Where("id = ?", cart.Product.SellerID).Select("average_rating").First(&seller).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve seller rating",
			})
			return
		}

		CartResponse = append(CartResponse, models.CartResponse{
			ProductID:    cart.ProductID,
			ProductName:  cart.Product.Name,
			CategoryID:   cart.Product.CategoryID,
			Description:  cart.Product.Description,
			Price:        cart.Product.Price,
			OfferAmount:  cart.Product.OfferAmount,
			FinalAmount:  RoundDecimalValue(finalAmount),
			Availability: cart.Product.Availability,
			Image:        cart.Product.Image,
			SellerRating: seller.AverageRating,
			ID:           cart.ID,
		})
	}

	formattedTotalAmount := fmt.Sprintf("%.2f", TotalAmount)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved all cart items",
		"data": gin.H{
			"cart":        CartResponse,
			"totalAmount": formattedTotalAmount,
			"itemCount":   ItemCount,
		},
	})
}

func RemoveItemFromCart(c *gin.Context) {
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

	cartIDStr := c.Query("cartid")
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid cartID",
		})
		return
	}

	var cart models.Cart

	if err := database.DB.First(&cart, cartID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "product is not present in the cart",
		})
		return
	}

	if cart.UserID != userIDStr {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "You are not authorized to remove this item from the cart.",
		})
		return
	}

	if err := database.DB.Delete(&cart, cartID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "unable to remove product from the cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully remove product from cart",
	})
}
