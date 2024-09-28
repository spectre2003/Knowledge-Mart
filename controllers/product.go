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
			"status":  false,
			"message": "seller not authorized",
		})
		return
	}

	sellerIDUint, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve seller information",
		})
		return
	}

	// Parse the request body
	var request models.AddProductRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process request",
		})
		return
	}

	// Validate the request
	validate := validator.New()
	if err := validate.Struct(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	// Check if product with the same name already exists for this seller
	var existingProduct models.Product
	if err := database.DB.Where("name = ? AND seller_id = ? AND deleted_at IS NULL", request.Name, sellerIDUint).First(&existingProduct).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
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
			"status":  false,
			"message": "failed to create product" + err.Error(),
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully added new product",
		"data": gin.H{
			"id":      newProduct.ID,
			"product": newProduct,
		},
	})
}

func EditProduct(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "seller not authorized",
		})
		return
	}

	_, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve seller information",
		})
		return
	}

	var Request models.EditProductRequest

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process request",
		})
		return
	}
	validate := validator.New()
	if err := validate.Struct(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	SellId := SellerIdbyProductId(Request.ProductID)

	if sellerID != SellId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request, product is not yours",
		})
		return
	}

	var existingProduct models.Product

	if err := database.DB.Where("id = ?", Request.ProductID).First(&existingProduct).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to fetch product from the database",
		})
		return
	}

	// Update the product fields
	existingProduct.Name = Request.Name
	existingProduct.Description = Request.Description
	existingProduct.Price = Request.Price
	existingProduct.Image = Request.Image
	existingProduct.Availability = Request.Availability

	// Use Select to update all fields including Availability
	if err := database.DB.Model(&existingProduct).Select("Name", "Description", "Price", "Image", "Availability").Updates(&existingProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully updated product information",
		"data": gin.H{
			"product": Request,
		},
	})
}

func DeleteProduct(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "seller not authorized",
		})
		return
	}

	_, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve seller information",
		})
		return
	}
	productIDStr := c.Query("productid")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "invalid product ID",
		})
		return
	}
	SellId := SellerIdbyProductId(uint(productID))

	if sellerID != SellId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request, product is not yours",
		})
		return
	}
	var product models.Product

	if err := database.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "product is not present in the database",
		})
		return
	}

	if err := database.DB.Delete(&product, productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "unable to delete the product from the database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully deleted the product",
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
			"status":  false,
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
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved products",
		"data": gin.H{
			"products": productResponse,
		},
	})
}
