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

	router.POST("/seller_login", controllers.SellerLogin)

	router.GET("/All_Products", controllers.ListAllProduct)

	userRoutes := router.Group("/user")
	userRoutes.Use(middleware.AuthRequired)
	{
		userRoutes.POST("/seller_registration", controllers.SellerRegister)
	}

	sellerRoutes := router.Group("/seller")
	sellerRoutes.Use(middleware.AuthRequired)
	{
		sellerRoutes.POST("/add_product", controllers.AddProduct)
		sellerRoutes.PATCH("/edit_product", controllers.EditProduct)
		sellerRoutes.DELETE("/delete_product", controllers.DeleteProduct)
	}

	adminRoutes := router.Group("/admin")
	adminRoutes.Use(middleware.AuthRequired)
	{
		adminRoutes.POST("/add_category", controllers.AddCatogory)
		adminRoutes.PATCH("/edit_category", controllers.EditCategory)
		adminRoutes.DELETE("/delete_category", controllers.DeleteCategory)
	}

}
