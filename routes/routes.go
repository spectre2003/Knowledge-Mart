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

	router.GET("/api/v1/googlelogin", controllers.GoogleHandleLogin)
	router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback)

	router.GET("/verifyemail/:email/:otp", controllers.VarifyEmail)
	router.POST("/resend_otp/:email", controllers.ResendOTP)

	router.POST("/seller_login", controllers.SellerLogin)

	router.GET("/All_Products", controllers.ListAllProduct)
	router.GET("/all_category", controllers.ListAllCategory)
	router.GET("/product_category", controllers.ListCategoryProductList)

	userRoutes := router.Group("/user")
	userRoutes.Use(middleware.AuthRequired)
	{
		userRoutes.POST("/seller_registration", controllers.SellerRegister)
		userRoutes.POST("/add_address", controllers.AddAddress)
		userRoutes.GET("/get_address", controllers.ListAllAddress)
		userRoutes.PUT("/edit_address", controllers.EditAddress)
		userRoutes.DELETE("/delete_address", controllers.DeleteAddress)
		userRoutes.GET("/user_profile", controllers.GetUserProfile)
		userRoutes.PUT("/edit_user_profile", controllers.EditUserProfile)
		userRoutes.PATCH("/edit_user_password", controllers.EditPassword)
	}

	sellerRoutes := router.Group("/seller")
	sellerRoutes.Use(middleware.AuthRequired)
	{
		sellerRoutes.POST("/add_product", controllers.AddProduct)
		sellerRoutes.PUT("/edit_product", controllers.EditProduct)
		sellerRoutes.DELETE("/delete_product", controllers.DeleteProduct)
		sellerRoutes.GET("/seller_profile", controllers.GetSellerProfile)
		sellerRoutes.PUT("/edit_seller", controllers.EditSellerProfile)
		sellerRoutes.GET("/product_by_seller", controllers.ListProductBySeller)
	}

	adminRoutes := router.Group("/admin")
	adminRoutes.Use(middleware.AuthRequired)
	{
		adminRoutes.POST("/add_category", controllers.AddCatogory)
		adminRoutes.PUT("/edit_category", controllers.EditCategory)
		adminRoutes.DELETE("/delete_category", controllers.DeleteCategory)
		adminRoutes.GET("/all_users", controllers.ListAllUsers)
		adminRoutes.GET("/list_blocked_users", controllers.ListBlockedUsers)
		adminRoutes.PATCH("/block_user", controllers.BlockUser)
		adminRoutes.PATCH("/unblock_user", controllers.UnBlockUser)
		adminRoutes.GET("/all_seller", controllers.ListAllSellers)
		adminRoutes.PATCH("/verify_seller", controllers.VerifySeller)
		adminRoutes.PATCH("/not_verified_seller", controllers.NotVerifySeller)
	}

}
