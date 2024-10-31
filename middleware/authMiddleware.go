package middleware

import (
	"fmt"
	"knowledgeMart/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthRequired(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	//fmt.Println("Authorization Header:", authHeader)

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Authorization header required",
		})
		c.Abort()
		return
	}
	tokenString := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Invalid token format",
		})
		c.Abort()
		return
	}

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Token validation failed: %v", err),
		})
		c.Abort()
		return
	}

	switch claims.Role {
	case "user":
		c.Set("userID", claims.ID)
	case "seller":
		c.Set("sellerID", claims.ID)
	case "admin":
		c.Set("adminID", claims.ID)
	default:
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "failed",
			"message": "Unauthorized role"})
		c.Abort()
		return
	}

	c.Next()
}
