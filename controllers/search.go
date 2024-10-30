package controllers

import (
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchProducts(c *gin.Context) {
	var products []models.Product
	var productResponse []models.ProductResponse

	categoryID := c.Query("category_id")
	sortBy := c.Query("sort_by")
	filterAvailable := c.Query("available")

	query := database.DB.Model(&products)

	if filterAvailable == "true" {
		query = query.Where("availability = ?", true)
	}

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	switch sortBy {
	case "price_asc":
		query = query.Order("offer_amount ASC")
	case "price_desc":
		query = query.Order("offer_amount DESC")
	case "newest":
		query = query.Order("created_at DESC")
	case "name_asc":
		query = query.Order("LOWER(name) ASC")
	case "name_desc":
		query = query.Order("LOWER(name) DESC")
	case "high_rating":
		query = query.Joins("JOIN sellers ON sellers.id = products.seller_id").Order("sellers.average_rating DESC")
	}

	tx := query.Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}

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
		"message": "successfully retrieved products",
		"data": gin.H{
			"products": productResponse,
		},
	})
}

func TopSellingProduct(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	sellerIDUint, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}

	type ProductSales struct {
		ProductID uint
		Count     int
	}

	var topProducts []ProductSales

	if err := database.DB.Table("order_items").
		Select("order_items.product_id, COUNT(*) as count").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.seller_id = ?", sellerIDUint).
		Group("order_items.product_id").
		Order("count DESC").
		Limit(10).
		Find(&topProducts).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve top-selling products",
		})
		return
	}

	var productResponse []models.ProductResponse
	for _, productSale := range topProducts {
		var product models.Product
		if err := database.DB.Where("id = ?", productSale.ProductID).First(&product).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "failed to retrieve product details",
			})
			return
		}

		var seller models.Seller
		if err := database.DB.Where("id = ?", product.SellerID).Select("average_rating").First(&seller).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "failed to retrieve seller rating",
			})
			return
		}

		fmt.Println(productSale.Count)

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
			//SalesCount:   productSale.Count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved top selling products",
		"data": gin.H{
			"products": productResponse,
		},
	})
}

func TopSellingCategory(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	sellerIDUint, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}
	type CategorySales struct {
		CategoryID   uint
		CategoryName string
		Count        int
	}

	var topCategories []CategorySales

	if err := database.DB.Table("order_items").
		Select("products.category_id, categories.name as category_name, COUNT(*) as count").
		Joins("JOIN products ON products.id = order_items.product_id").
		Joins("JOIN categories ON categories.id = products.category_id").
		Where("products.seller_id = ?", sellerIDUint).
		Group("products.category_id, categories.name").
		Order("count DESC").
		Limit(10).
		Find(&topCategories).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve top-selling categories",
		})
		return
	}

	var bestSellingCategoryID uint
	var bestSellingCategoryName string
	var maxSales int
	if len(topCategories) > 0 {
		bestSellingCategoryID = topCategories[0].CategoryID
		bestSellingCategoryName = topCategories[0].CategoryName
		maxSales = topCategories[0].Count
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved best-selling category",
		"data": gin.H{
			"top_categories": topCategories,
			"best_selling_category": gin.H{
				"category_id":   bestSellingCategoryID,
				"category_name": bestSellingCategoryName,
				"sales_count":   maxSales,
			},
		},
	})
}
