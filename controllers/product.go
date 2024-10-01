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
	// Retrieve sellerID from context and check for its existence
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

	// Parse the request body
	var request models.AddProductRequest

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
		return
	}

	// Validate the request
	validate := validator.New()
	if err := validate.Struct(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	// Check if product with the same name already exists for this seller
	var existingProduct models.Product

	if err := database.DB.Where("name = ? AND seller_id = ? AND deleted_at IS NULL", request.Name, sellerIDUint).First(&existingProduct).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "product with the same name already exists for this seller",
		})
		return
	}

	// Create a new product
	newProduct := models.Product{
		SellerID:     sellerID.(uint),
		CategoryID:   request.CategoryID,
		Name:         request.Name,
		Description:  request.Description,
		Price:        request.Price,
		Image:        request.Image,
		Availability: true, // Assuming products are available by default
	}

	// Save the new product in the database
	if err := database.DB.Create(&newProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to create product" + err.Error(),
		})
		return
	}

	// Respond with success
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
			"image":        newProduct.Image,
			"availability": newProduct.Availability,
		},
	})
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
		existingProduct.Price = Request.Price
	}

	if len(Request.Image) > 0 {
		existingProduct.Image = Request.Image
	}

	if Request.Availability != nil {
		existingProduct.Availability = *Request.Availability
	}

	// Use Select to update all fields including Availability
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
			"image":        existingProduct.Image,
			"availability": existingProduct.Availability,
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

	_, ok := sellerID.(uint)
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

	if sellerID != SellId {
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
		productResponse = append(productResponse, models.ProductResponse{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Image:        product.Image,
			Availability: product.Availability,
			SellerID:     product.SellerID,
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
