package controllers

import (
	"errors"
	"knowledgeMart/config"
	"knowledgeMart/models"
	"knowledgeMart/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminLogin(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	var admin models.Admin
	if tx := database.DB.Where("email = ?", loginData.Email).First(&admin); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Email not present in the admin table",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}

	if admin.Password != loginData.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}
	token, err := utils.GenerateJWT(admin.ID, "admin")
	if token == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "login success",
	})
	//c.JSON(http.StatusSeeOther, gin.H{"redirect": "/admin-panel"})

}
