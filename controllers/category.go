package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ListAllCategory(c *gin.Context) {
	var Categories []models.Category

	tx := database.DB.Select("*").Find(&Categories)

	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrieve data from the database, or the product doesn't exist",
		})
		return
	}

	var categoryResponse []models.CatgoryResponse

	for _, category := range Categories {
		categoryResponse = append(categoryResponse, models.CatgoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			Image:       category.Image,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved products",
		"data": gin.H{
			"categories": categoryResponse,
		},
	})
}

// func ListCategoryProductList (c *gin.Context){
// 	var  categories []models.Category

// }

func AddCatogory(c *gin.Context) {
	var Request models.AddCategoryRequest

	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve admin information",
		})
		return
	}

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process incoming request",
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
	words := strings.Fields(Request.Description)
	wordCount := len(words)

	if wordCount < 10 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "description must be a minimum of 10 words",
		})
		return
	}

	var existCategory models.Category

	if err := database.DB.Where("name = ?", Request.Name).First(&existCategory).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "category already exists",
		})
		return
	}

	NewCategory := models.Category{
		Name:        Request.Name,
		Description: Request.Description,
		Image:       Request.Image,
	}

	if err := database.DB.Save(&NewCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "unable to add new category, server error ",
		})
		return

	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successully added a new category",
		"data": gin.H{
			"category": Request,
		},
	})

}

func EditCategory(c *gin.Context) {
	var Request models.EditCategoryRequest
	var existCategory models.Category

	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve admin information",
		})
		return
	}
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process incoming request",
		})
		return
	}

	if err := database.DB.First(&existCategory, Request.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "category not found",
		})
		return
	}

	if Request.Name != existCategory.Name {
		existCategory.Name = Request.Name
	}
	existCategory.Description = Request.Description
	existCategory.Image = Request.Image

	if err := database.DB.Save(&existCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update category details",
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully updated category",
	})
}

func DeleteCategory(c *gin.Context) {
	// Retrieve the adminID to verify authorization
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "not authorized",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retrieve admin information",
		})
		return
	}

	// Retrieve the categoryID from query parameters
	categoryIDStr := c.Query("categoryid")
	if categoryIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "categoryid is required",
		})
		return
	}

	var category models.Category
	// Fetch the category using the categoryID
	if err := database.DB.First(&category, categoryIDStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch category from the database",
		})
		return
	}

	// Check if there are products associated with this category
	var productCount int64
	result := database.DB.Model(&models.Product{}).Where("category_id = ?", category.ID).Count(&productCount)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to check products for this category",
		})
		return
	}

	if productCount > 0 {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"status":  false,
			"message": "category contains products, change the category of these products before using this endpoint",
		})
		return
	}

	// Delete the category from the database
	if err := database.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to delete category from the database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully deleted category from the database",
	})
}
