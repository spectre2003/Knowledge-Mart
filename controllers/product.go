package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func AddProduct(c *gin.Context) {
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

	var request models.AddProductRequest

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var category models.Category
	if err := database.DB.First(&category, request.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid category ID",
		})
		return
	}

	var existingProduct models.Product

	if err := database.DB.Where("name = ? AND seller_id = ? AND deleted_at IS NULL", request.Name, sellerIDUint).First(&existingProduct).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "product with the same name already exists for this seller",
		})
		return
	}

	if request.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "product price must be a positive integer",
		})
		return
	}

	finalAmount := calculateFinalAmount(request.OfferAmount, category.OfferPercentage)

	newProduct := models.Product{
		SellerID:     sellerID.(uint),
		CategoryID:   request.CategoryID,
		Name:         request.Name,
		Description:  request.Description,
		Price:        request.Price,
		OfferAmount:  request.OfferAmount,
		Image:        request.Image,
		Availability: true,
	}
	if err := database.DB.Create(&newProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to create product" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully added new product",
		"data": gin.H{
			"id":           newProduct.ID,
			"product_name": newProduct.Name,
			"category_id":  newProduct.CategoryID,
			"seller_id":    newProduct.SellerID,
			"describtion":  newProduct.Description,
			"price":        newProduct.Price,
			"offer_amount": newProduct.OfferAmount,
			"final_amount": finalAmount,
			"image":        newProduct.Image,
			"availability": newProduct.Availability,
		},
	})
}

func calculateFinalAmount(offerAmount float64, offerPercentage uint) float64 {
	if offerPercentage > 0 {
		discount := (offerAmount * float64(offerPercentage)) / 100
		finalAmount := offerAmount - discount
		if finalAmount < 0 {
			finalAmount = 0
		}
		return finalAmount
	}
	return offerAmount
}

func EditProduct(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	_, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}

	var Request models.EditProductRequest

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

	SellId := SellerIdbyProductId(Request.ProductID)

	if sellerID != SellId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "unauthorized request, product is not yours",
		})
		return
	}

	var existingProduct models.Product

	if err := database.DB.Where("id = ?", Request.ProductID).First(&existingProduct).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to fetch product from the database",
		})
		return
	}

	// Update only if each field in the request is non-empty or non-zero
	if Request.Name != "" {
		existingProduct.Name = Request.Name
	}

	if Request.Description != "" {
		existingProduct.Description = Request.Description
	}

	if Request.Price != 0 {
		if Request.Price <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "price must be a positive number",
			})
			return
		}
		existingProduct.Price = Request.Price
	}

	if Request.OfferAmount != 0 {
		if Request.OfferAmount <= 0 && Request.OfferAmount < Request.Price {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "price must be a positive number",
			})
			return
		}
		existingProduct.OfferAmount = Request.OfferAmount
	}

	if len(Request.Image) > 0 {
		existingProduct.Image = Request.Image
	}

	if Request.Availability != nil {
		existingProduct.Availability = *Request.Availability
	}

	if Request.CategoryID != 0 {

		var category models.Category
		if err := database.DB.First(&category, Request.CategoryID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "invalid category ID",
			})
			return
		}

		existingProduct.CategoryID = Request.CategoryID
	}
	if err := database.DB.Model(&existingProduct).Updates(existingProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully updated product information",
		"data": gin.H{
			"id":           existingProduct.ID,
			"name":         existingProduct.Name,
			"description":  existingProduct.Description,
			"price":        existingProduct.Price,
			"offer_amount": existingProduct.OfferAmount,
			"image":        existingProduct.Image,
			"availability": existingProduct.Availability,
			"categoryID":   existingProduct.CategoryID,
		},
	})
}

func DeleteProduct(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	sellerIDStr, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}
	productIDStr := c.Query("productid")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid product ID",
		})
		return
	}
	SellId := SellerIdbyProductId(uint(productID))

	if sellerIDStr != SellId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "unauthorized request, product is not yours",
		})
		return
	}
	var product models.Product

	if err := database.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "product is not present in the database",
		})
		return
	}

	if err := database.DB.Delete(&product, productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "unable to delete the product from the database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully deleted the product",
	})
}

func ListProductBySeller(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	_, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}

	sellID := c.Query("id")
	if sellID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "missing category id",
		})
		return
	}
	var products []models.Product
	tx := database.DB.Select("*").Where("seller_id = ?", sellID).Find(&products)

	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve products for the specified seller",
		})
		return
	}

	var productResponse []models.ProductResponse

	for _, product := range products {
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Price:        product.Price,
			OfferAmount:  product.OfferAmount,
			Description:  product.Description,
			Image:        product.Image,
			Availability: product.Availability,
			CategoryID:   product.CategoryID,
			SellerID:     product.SellerID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved products for category",
		"data": gin.H{
			"products": productResponse,
		},
	})
}

func SellerIdbyProductId(ProductId uint) uint {
	var Product models.Product
	if err := database.DB.Where("id = ?", ProductId).First(&Product).Error; err != nil {
		return 0
	}
	return Product.SellerID
}

func ListAllProduct(c *gin.Context) {
	var Products []models.Product

	tx := database.DB.Select("*").Find(&Products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the database, or the product doesn't exist",
		})
		return
	}
	var productResponse []models.ProductResponse

	for _, product := range Products {
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

func AddProductOffer(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	_, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}

	var request models.AddOfferRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid request body ",
		})
		return
	}

	var product models.Product

	if err := database.DB.Where("id = ?", request.ProductID).First(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to find the product",
		})
		return
	}

	product.OfferAmount = request.OfferAmount

	if err := database.DB.Model(&product).Where("id = ?", request.ProductID).Update("offer_amount", request.OfferAmount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to add the offer amount",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully added the offer amount",
		"data":    product,
	})
}
