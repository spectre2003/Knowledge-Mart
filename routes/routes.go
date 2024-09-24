package routes

import (
	"knowledgeMart/controllers"
	"knowledgeMart/middleware"

	"github.com/gin-gonic/gin"
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

	protected := router.Group("/user")
	protected.Use(middleware.AuthRequired)
	{
		protected.POST("/seller_registration", controllers.SellerRegister)
		//protected.GET("/profile", controllers.UserProfile)
	}
}
