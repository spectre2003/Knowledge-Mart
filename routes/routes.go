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

	//admin auth
	router.POST("/admin_login", controllers.AdminLogin)

	//user auth
	router.POST("/user_signup", controllers.EmailSignup)
	router.POST("/user_login", controllers.EmailLogin)

	//user google auth
	router.GET("/api/v1/googlelogin", controllers.GoogleHandleLogin)
	router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback)

	//email varification
	router.GET("/verifyemail/:email/:otp", controllers.VarifyEmail)
	router.POST("/resend_otp/:email", controllers.ResendOTP)

	//seller auth
	router.POST("/seller_login", controllers.SellerLogin)

	//products
	router.GET("/All_Products", controllers.ListAllProduct)
	router.GET("/all_category", controllers.ListAllCategory)
	router.GET("/product_category", controllers.ListCategoryProductList)
	router.GET("/product_price_lowtohigh", controllers.SearchProductLtoH)
	router.GET("/product_price_hightolow", controllers.SearchProductHtoL)
	router.GET("/product_new", controllers.SearchProductNew)
	router.GET("/product_a_to_z", controllers.SearchProductAtoZ)
	router.GET("/product_z_to_a", controllers.SearchProductZtoA)
	router.GET("/product_popularity", controllers.SearchProductHighRatedFirst)

	userRoutes := router.Group("/user")
	userRoutes.Use(middleware.AuthRequired)
	{
		userRoutes.POST("/seller_registration", controllers.SellerRegister)

		//address
		userRoutes.POST("/add_address", controllers.AddAddress)
		userRoutes.GET("/get_address", controllers.ListAllAddress)
		userRoutes.PUT("/edit_address", controllers.EditAddress)
		userRoutes.DELETE("/delete_address", controllers.DeleteAddress)

		//profile
		userRoutes.GET("/user_profile", controllers.GetUserProfile)
		userRoutes.PUT("/edit_user_profile", controllers.EditUserProfile)
		userRoutes.PATCH("/edit_user_password", controllers.EditPassword)

		//cart
		userRoutes.POST("/add_to_cart", controllers.AddToCart)
		userRoutes.GET("/cart_view", controllers.ListAllCart)
		userRoutes.DELETE("/remove_cart", controllers.RemoveItemFromCart)

		//order
		userRoutes.POST("/order_place", controllers.PlaceOrder)
		userRoutes.GET("/my_orders", controllers.UserCheckOrderStatus)
		userRoutes.PATCH("/order_cancel", controllers.CancelOrder)

		//note sharing
		userRoutes.POST("/file_upload", controllers.UploadFile)

		//rating
		userRoutes.POST("/seller_rating", controllers.SellerRating)

		//whishlist
		userRoutes.POST("/add_to_whishlist", controllers.AddToWhishList)
		userRoutes.GET("/whishlist_view", controllers.ListAllWhishList)
		userRoutes.DELETE("/remove_whishlist", controllers.RemoveItemFromwhishlist)
	}

	sellerRoutes := router.Group("/seller")
	sellerRoutes.Use(middleware.AuthRequired)
	{
		//products
		sellerRoutes.POST("/add_product", controllers.AddProduct)
		sellerRoutes.PUT("/edit_product", controllers.EditProduct)
		sellerRoutes.DELETE("/delete_product", controllers.DeleteProduct)

		//profile
		sellerRoutes.GET("/seller_profile", controllers.GetSellerProfile)
		sellerRoutes.PUT("/edit_seller", controllers.EditSellerProfile)
		sellerRoutes.PATCH("/edit_seller_password", controllers.EditSellerPassword)
		sellerRoutes.GET("/product_by_seller", controllers.ListProductBySeller)

		//order
		sellerRoutes.GET("/order_list", controllers.GetUserOrders)
		sellerRoutes.PATCH("/update_order_status", controllers.SellerUpdateOrderStatus)
		sellerRoutes.PATCH("/order_cancel", controllers.CancelOrder)
	}

	adminRoutes := router.Group("/admin")
	adminRoutes.Use(middleware.AuthRequired)
	{
		//category
		adminRoutes.POST("/add_category", controllers.AddCatogory)
		adminRoutes.PUT("/edit_category", controllers.EditCategory)
		adminRoutes.DELETE("/delete_category", controllers.DeleteCategory)

		//user management
		adminRoutes.GET("/all_users", controllers.ListAllUsers)
		adminRoutes.GET("/list_blocked_users", controllers.ListBlockedUsers)
		adminRoutes.PATCH("/block_user", controllers.BlockUser)
		adminRoutes.PATCH("/unblock_user", controllers.UnBlockUser)

		//seller management
		adminRoutes.GET("/all_seller", controllers.ListAllSellers)
		adminRoutes.PATCH("/verify_seller", controllers.VerifySeller)
		adminRoutes.PATCH("/not_verified_seller", controllers.NotVerifySeller)
	}

}
