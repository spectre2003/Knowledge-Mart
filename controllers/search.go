package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchProductLtoH(c *gin.Context) {
	var products []models.Product

	tx := database.DB.Where("availability = ?", true).Order("offer_amount ASC").Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}

	var productResponse []models.ProductResponse

	for _, product := range products {
		var seller models.Seller
		if err := database.DB.Where("id = ?", product.SellerID).Select("average_rating").First(&seller).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve seller rating",
			})
			return
		}
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			OfferAmount:  product.OfferAmount,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
			SellerRating: seller.AverageRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved products sorted by price",
		"data": gin.H{
			"products": productResponse,
		},
	})
}

func SearchProductHtoL(c *gin.Context) {
	var products []models.Product

	tx := database.DB.Where("availability = ?", true).Order("offer_amount DESC").Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}

	var productResponse []models.ProductResponse

	for _, product := range products {
		var seller models.Seller
		if err := database.DB.Where("id = ?", product.SellerID).Select("average_rating").First(&seller).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve seller rating",
			})
			return
		}
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			OfferAmount:  product.OfferAmount,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
			SellerRating: seller.AverageRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved products sorted by price",
		"data": gin.H{
			"products": productResponse,
		},
	})
}

func SearchProductNew(c *gin.Context) {
	var products []models.Product

	tx := database.DB.Where("availability = ?", true).Order("products.created_at DESC").Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}

	var productResponse []models.ProductResponse

	for _, product := range products {
		var seller models.Seller
		if err := database.DB.Where("id = ?", product.SellerID).Select("average_rating").First(&seller).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve seller rating",
			})
			return
		}
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			OfferAmount:  product.OfferAmount,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
			SellerRating: seller.AverageRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved products sorted by new arrival",
		"data": gin.H{
			"products": productResponse,
		},
	})
}

func SearchProductAtoZ(c *gin.Context) {
	var products []models.Product

	tx := database.DB.Where("availability = ?", true).Order("LOWER(products.name) ASC").Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}

	var productResponse []models.ProductResponse

	for _, product := range products {
		var seller models.Seller
		if err := database.DB.Where("id = ?", product.SellerID).Select("average_rating").First(&seller).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve seller rating",
			})
			return
		}
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			OfferAmount:  product.OfferAmount,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
			SellerRating: seller.AverageRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved products sorted by alphabetic order",
		"data": gin.H{
			"products": productResponse,
		},
	})
}

func SearchProductZtoA(c *gin.Context) {
	var products []models.Product

	tx := database.DB.Where("availability = ?", true).Order("LOWER(products.name) DESC").Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}

	var productResponse []models.ProductResponse

	for _, product := range products {
		var seller models.Seller
		if err := database.DB.Where("id = ?", product.SellerID).Select("average_rating").First(&seller).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve seller rating",
			})
			return
		}
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			OfferAmount:  product.OfferAmount,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
			SellerRating: seller.AverageRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved products sorted by reverce alphbetic order",
		"data": gin.H{
			"products": productResponse,
		},
	})
}

func SearchProductHighRatedFirst(c *gin.Context) {
	var products []models.Product

	tx := database.DB.
		Joins("JOIN sellers ON sellers.id = products.seller_id").
		Where("products.availability = ?", true).
		Order("sellers.average_rating DESC").
		Find(&products)

	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}

	var productResponse []models.ProductResponse

	for _, product := range products {
		var seller models.Seller
		if err := database.DB.Where("id = ?", product.SellerID).Select("average_rating").First(&seller).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to retrieve seller rating",
			})
			return
		}

		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			OfferAmount:  product.OfferAmount,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
			SellerRating: seller.AverageRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved products sorted by seller's high ratings",
		"data": gin.H{
			"products": productResponse,
		},
	})
}
