package routes

import (
	"knowledgeMart/controllers"
	"knowledgeMart/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {

	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server status ok",
		})
	})

	//admin auth
	router.POST("/api/v1/admin/login", controllers.AdminLogin)

	//user auth
	router.POST("/api/v1/user/signup", controllers.EmailSignup)
	router.POST("/api/v1/user/login", controllers.EmailLogin)

	//user google auth
	router.GET("/api/v1/googlelogin", controllers.GoogleHandleLogin)
	router.GET("/api/v1/googlecallback", controllers.GoogleHandleCallback)

	//email varification
	router.GET("/api/v1/verifyemail/:email/:otp", controllers.VarifyEmail)
	router.POST("/api/v1/otp/resend/:email", controllers.ResendOTP)

	//seller auth
	router.POST("/api/v1/seller/login", controllers.SellerLogin)

	//products search
	router.GET("/api/v1/public/product/search", controllers.SearchProducts)
	router.GET("/api/v1/public/category/all", controllers.ListAllCategory)

	//coupon
	router.GET("/api/v1/public/coupon/all", controllers.GetAllCoupons)

	//course & notes
	router.GET("/api/v1/public/course/all", controllers.GetAllCoursesWithDetails)
	router.GET("/api/v1/public/notes/all", controllers.GetAllNotes)

	userRoutes := router.Group("/api/v1/user")
	userRoutes.Use(middleware.AuthRequired)
	{
		userRoutes.POST("/seller/registration", controllers.SellerRegister)

		//address
		userRoutes.POST("/address/create", controllers.AddAddress)
		userRoutes.GET("/address/get", controllers.ListAllAddress)
		userRoutes.PUT("/address/edit", controllers.EditAddress)
		userRoutes.DELETE("/address/delete", controllers.DeleteAddress)

		//profile
		userRoutes.GET("/profile", controllers.GetUserProfile)
		userRoutes.PUT("/profile/edit", controllers.EditUserProfile)
		userRoutes.PATCH("/password/edit", controllers.EditPassword)

		//cart
		userRoutes.POST("/cart/add", controllers.AddToCart)
		userRoutes.GET("/cart/view", controllers.ListAllCart)
		userRoutes.DELETE("/cart/remove", controllers.RemoveItemFromCart)
		userRoutes.GET("/coupon/cart", controllers.ApplyCouponOnCart)

		//order
		userRoutes.POST("/order/create", controllers.PlaceOrder)
		userRoutes.GET("/order/check", controllers.UserCheckOrderStatus)
		userRoutes.PATCH("/order/cancel", controllers.CancelOrder)
		userRoutes.PATCH("/order/return", controllers.ReturnOrder)
		userRoutes.GET("/order/invoice", controllers.OrderInvoice)

		//note sharing
		userRoutes.POST("/file/upload", controllers.UploadFile)
		userRoutes.POST("/note/upload", controllers.UploadNote)
		userRoutes.PATCH("/note/edit", controllers.EditNote)
		userRoutes.DELETE("/note/delete", controllers.DeleteNote)
		userRoutes.GET("/note/view", controllers.GetUserNotes)

		//rating
		userRoutes.POST("/seller-rating", controllers.SellerRating)

		//whishlist
		userRoutes.POST("/whishlist/add", controllers.AddToWhishList)
		userRoutes.GET("/whishlist/view", controllers.ListAllWhishList)
		userRoutes.DELETE("/whishlist/remove", controllers.RemoveItemFromwhishlist)

		//wallet history
		userRoutes.GET("/wallet/history", controllers.GetUserWalletHistory)

	}
	//razorpay
	router.POST("/verify-payment/:orderID", controllers.VerifyPayment)
	router.POST("/payment-failed/:orderID", controllers.HandleFailedPayment)
	router.GET("/payment-method", controllers.RenderRazorpay)
	router.POST("/create-order/:orderID", controllers.CreateOrder)
	router.GET("/check-failed-attempts/:orderID", controllers.CheckFailedAttempts)

	sellerRoutes := router.Group("/api/v1/seller")
	sellerRoutes.Use(middleware.AuthRequired)
	{
		//products
		sellerRoutes.POST("/product/add", controllers.AddProduct)
		sellerRoutes.PUT("/product/edit", controllers.EditProduct)
		sellerRoutes.DELETE("/product/delete", controllers.DeleteProduct)
		sellerRoutes.GET("/product/view", controllers.ListProductBySeller)

		//profile
		sellerRoutes.GET("/profile", controllers.GetSellerProfile)
		sellerRoutes.PUT("/profile/edit", controllers.EditSellerProfile)
		sellerRoutes.PATCH("/password/edit", controllers.EditSellerPassword)

		//order
		sellerRoutes.GET("/order/view", controllers.GetUserOrders)
		sellerRoutes.PATCH("/order/status/update", controllers.SellerUpdateOrderStatus)
		sellerRoutes.PATCH("/order/status/cancel", controllers.CancelOrder)

		//sales report
		sellerRoutes.GET("/report/all", controllers.SellerOverAllSalesReport)
		sellerRoutes.GET("/report/download/pdf", controllers.DownloadSalesReportPDF)
		sellerRoutes.GET("/report/download/excel", controllers.DownloadSalesReportExcel)

		//wallet history
		sellerRoutes.GET("/wallet/history", controllers.GetSellerWalletHistory)

		//top selling
		sellerRoutes.GET("/product/top-selling", controllers.TopSellingProduct)
		sellerRoutes.GET("/category/top-selling", controllers.TopSellingCategory)

	}

	adminRoutes := router.Group("/api/v1/admin")
	adminRoutes.Use(middleware.AuthRequired)
	{
		//category
		adminRoutes.POST("/category/add", controllers.AddCatogory)
		adminRoutes.PUT("/category/edit", controllers.EditCategory)
		adminRoutes.DELETE("/category/delete", controllers.DeleteCategory)

		//user management
		adminRoutes.GET("/view/users", controllers.ListAllUsers)
		adminRoutes.GET("/view/blocked-users", controllers.ListBlockedUsers)
		adminRoutes.PATCH("/block/user", controllers.BlockUser)
		adminRoutes.PATCH("/unblock/user", controllers.UnBlockUser)

		//seller management
		adminRoutes.GET("/view/sellers", controllers.ListAllSellers)
		adminRoutes.PATCH("/verify/seller", controllers.VerifySeller)
		adminRoutes.PATCH("/un-verify/seller", controllers.NotVerifySeller)

		//Coupon management
		adminRoutes.POST("/coupon/create", controllers.CreateCoupen)
		adminRoutes.PATCH("/coupon/update", controllers.UpdateCoupon)
		adminRoutes.DELETE("/coupon/delete", controllers.DeleteCoupon)

		//course management
		adminRoutes.POST("/course/create", controllers.CreateCourse)
		adminRoutes.PATCH("/course/edit", controllers.EditCourse)
		adminRoutes.DELETE("/course/delete", controllers.DeleteCourse)

		//semester management
		adminRoutes.POST("/semester/create", controllers.CreateSemester)
		adminRoutes.PATCH("/semester/edit", controllers.EditSemester)
		adminRoutes.DELETE("/semester/delete", controllers.DeleteSemester)

		//subject management
		adminRoutes.POST("/subject/create", controllers.CreateSubject)
		adminRoutes.PATCH("/subject/edit", controllers.EditSubject)
		adminRoutes.DELETE("/subject/delete", controllers.DeleteSubject)

	}

}
