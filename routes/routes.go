package routes

import (
	"github.com/gin-gonic/gin"
	"knowledgeMart/controllers"
)

func RegisterRoutes(router *gin.Engine) {

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server status ok",
		})
	})
	router.POST("/admin_login", controllers.AdminLogin)

	router.POST("/user_signup", controllers.EmailSignup)
	router.POST("/user_login", controllers.EmailLogin)

	router.GET("/verifyemail/:email/:otp", controllers.VarifyEmail)
}
