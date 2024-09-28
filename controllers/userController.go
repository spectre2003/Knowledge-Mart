package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListAllUsers(c *gin.Context) {
	// Check if admin is authorized
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

	// Fetch all users from the database
	var userResponse []models.UserResponse
	var users []models.User

	tx := database.DB.Find(&users)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	// Prepare the response by iterating over the fetched users
	for _, user := range users {
		userResponse = append(userResponse, models.UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Picture:     user.Picture,
			//Address:     user.Address,
			Blocked:    user.Blocked,
			IsVerified: user.IsVerified,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved user information",
		"data": gin.H{
			"users": userResponse,
		},
	})
}

func BlockUser(c *gin.Context) {
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

	userId := c.Query("userid")

	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "userid is required",
		})
		return
	}

	var user models.User

	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch user from the database",
		})
		return
	}

	if user.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":  false,
			"message": "user is already blocked",
		})
		return
	}

	user.Blocked = true

	tx := database.DB.Model(&user).Update("blocked", user.Blocked)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to change the block status ",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  false,
		"message": "successfully blocked the user",
	})

}

func UnBlockUser(c *gin.Context) {
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

	userId := c.Query("userid")

	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "userid is required",
		})
		return
	}

	var user models.User

	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch user from the database",
		})
		return
	}

	if !user.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status":  false,
			"message": "user is already unblocked",
		})
		return
	}

	user.Blocked = false

	tx := database.DB.Model(&user).Update("blocked", user.Blocked)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to change the block status ",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  false,
		"message": "successfully unblocked the user",
	})

}

func ListBlockedUsers(c *gin.Context) {
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

	var blockedUser []models.UserResponse
	var users []models.User

	tx := database.DB.Where("deleted_at IS NULL AND blocked = ?", true).Find(&users)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrieve blocked user data from the database, or the data doesn't exists",
		})
		return
	}

	for _, user := range users {
		blockedUser = append(blockedUser, models.UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Picture:     user.Picture,
			//Address:     user.Address,
			Blocked:    user.Blocked,
			IsVerified: user.IsVerified,
		})
	}

	if len(blockedUser) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "no blocked users found",
			"data":    blockedUser,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved blocked user's data",
		"data": gin.H{
			"blocked_users": blockedUser,
		},
	})
}
