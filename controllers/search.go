package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchProductLtoH(c *gin.Context) {
	var products []models.Product

	tx := database.DB.Where("availability = ?", true).Order("price ASC").Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}

	var productResponse []models.ProductResponse

	for _, product := range products {
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
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

	tx := database.DB.Where("availability = ?", true).Order("price DESC").Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}

	var productResponse []models.ProductResponse

	for _, product := range products {
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
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
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
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
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
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
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
			CategoryID:   product.CategoryID,
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

// func SearchProductLtoH(c *gin.Context) {
// 	var productResponse []models.ProductResponse
// 	err := database.DB.Model(&models.Product{}).
// 		Select("products.id, products.name, products.description, products.image,products.price,products.availability,categories.name AS category_name,products.seller_id,products.category_id").
// 		Joins("JOIN categories ON categories.id=products.category_id").
// 		Order("products.price ASC").Find(&productResponse)
// 	if err.Error != nil {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"message": "failed to retrieve data from the products database, or the data doesn't exist",
// 			"error":   err.Error,
// 		})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"products": productResponse,
// 	})
// }

// func Search_P_HtoL(c *gin.Context) {
// 	var products []model.ViewProductList
// 	ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("price DESC").Find(&products)
// 	if ty.Error != nil {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"message": "failed to retrieve data from the products database, or the data doesn't exist",
// 			"error":   ty.Error,
// 		})
// 		return
// 	}
// 	for _, val := range products {
// 		c.JSON(http.StatusOK, gin.H{
// 			"products": val,
// 		})
// 	}
// }

// func SearchNew(c *gin.Context) {
// 	var products []model.ViewProductList
// 	ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("products.created_at DESC").Find(&products)
// 	if ty.Error != nil {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"message": "failed to retrieve data from the products database, or the data doesn't exist",
// 			"error":   ty.Error,
// 		})
// 		return
// 	}
// 	for _, val := range products {
// 		c.JSON(http.StatusOK, gin.H{
// 			"products": val,
// 		})
// 	}
// }

// func SearchAtoZ(c *gin.Context) {
// 	var products []model.ViewProductList
// 	ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("LOWER(products.name) ASC").Find(&products)
// 	if ty.Error != nil {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"message": "failed to retrieve data from the products database, or the data doesn't exist",
// 			"error":   ty.Error,
// 		})
// 		return
// 	}
// 	for _, val := range products {
// 		c.JSON(http.StatusOK, gin.H{
// 			"products": val,
// 		})
// 	}
// }

// func SearchZtoA(c *gin.Context) {
// 	var products []model.ViewProductList
// 	ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("LOWER(products.name) DESC").Find(&products)
// 	if ty.Error != nil {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"message": "failed to retrieve data from the products database, or the data doesn't exist",
// 			"error":   ty.Error,
// 		})
// 		return
// 	}
// 	for _, val := range products {
// 		c.JSON(http.StatusOK, gin.H{
// 			"products": val,
// 		})
// 	}
// }
