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
	fmt.Println("Authorization Header:", authHeader)

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization header required"})
		c.Abort()
		return
	}
	tokenString := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		c.Abort()
		return
	}

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		c.Abort()
		return
	}
	fmt.Println(tokenString)
	fmt.Println("Extracted UserID:", claims.UserID)

	c.Set("userID", claims.UserID)

	c.Next()
}
