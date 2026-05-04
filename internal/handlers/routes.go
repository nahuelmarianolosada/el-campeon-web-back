package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/middleware"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/cart"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/order"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/payment"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/product"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/product/variant"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/report"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/user"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Inicializar repositorios
	userRepo := repositories.NewUserRepository(db)
	productRepo := repositories.NewProductRepository(db)
	cartRepo := repositories.NewCartRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	reportRepo := repositories.NewReportRepository(db)
	variantRepo := repositories.NewProductVariantRepository(db)

	// Inicializar servicios
	userService := user.NewUserService(userRepo, cfg)
	productService := product.NewProductService(productRepo)
	cartService := cart.NewCartService(cartRepo, productRepo)
	orderService := order.NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo)
	paymentService := payment.NewPaymentService(paymentRepo, orderRepo, cfg)
	reportService := report.NewReportService(reportRepo)
	variantService := variant.NewProductVariantService(variantRepo, productRepo)

	// Inicializar handlers
	authHandler := NewAuthHandler(userService)
	productHandler := NewProductHandler(productService)
	cartHandler := NewCartHandler(cartService)
	orderHandler := NewOrderHandler(orderService)
	paymentHandler := NewPaymentHandler(paymentService)
	reportHandler := NewReportHandler(reportService)
	variantHandler := NewProductVariantHandler(variantService)

	// Auth routes (sin autenticación)
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}

	// Public product routes
	productRoutes := router.Group("/api/products")
	{
		productRoutes.GET("", productHandler.ListProducts)
		productRoutes.GET("/:id", productHandler.GetProduct)
		productRoutes.GET("/sku", productHandler.GetProductBySKU)
		productRoutes.GET("/category/:category", productHandler.ListProductsByCategory)
		productRoutes.GET("/:id/variants", variantHandler.GetProductVariants)
		productRoutes.GET("/:id/variant-combinations", variantHandler.GetProductVariantCombinations)
	}

	// Public variant combination routes
	variantRoutes := router.Group("/api/variant-combinations")
	{
		variantRoutes.GET("/:combinationId", variantHandler.GetVariantCombination)
		variantRoutes.GET("/sku", variantHandler.GetVariantCombinationBySKU)
	}

	// Public variant routes
	variantSingleRoutes := router.Group("/api/variants")
	{
		variantSingleRoutes.GET("/:variantId", variantHandler.GetVariant)
	}

	// Protected routes - requieren autenticación
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// Cart routes
		cartGroup := protected.Group("/api/cart")
		{
			cartGroup.GET("", cartHandler.GetCart)
			cartGroup.POST("/items", cartHandler.AddToCart)
			cartGroup.PUT("/items/:itemId", cartHandler.UpdateCartItem)
			cartGroup.DELETE("/items/:itemId", cartHandler.RemoveFromCart)
			cartGroup.DELETE("", cartHandler.ClearCart)
			cartGroup.GET("/total", cartHandler.GetCartTotal)
		}

		// Order routes
		orderGroup := protected.Group("/api/orders")
		{
			orderGroup.POST("", orderHandler.CreateOrder)
			orderGroup.GET("/my", orderHandler.GetMyOrders)
			orderGroup.GET("/:id", orderHandler.GetOrder)
		}

		// Payment routes
		paymentGroup := protected.Group("/api/payments")
		{
			paymentGroup.POST("", paymentHandler.CreatePayment)
			paymentGroup.GET("/my", paymentHandler.GetMyPayments)
			paymentGroup.GET("/:id", paymentHandler.GetPayment)
			paymentGroup.GET("/order/:orderId", paymentHandler.GetPaymentByOrderID)
		}
	}

	// Admin routes - requieren autenticación y rol ADMIN
	admin := router.Group("")
	admin.Use(middleware.AuthMiddleware(cfg))
	admin.Use(middleware.AdminMiddleware())
	{
		// Order admin routes
		userAdmin := admin.Group("/api/admin")
		{
			userAdmin.POST("/users", authHandler.RegisterAdmin)
		}

		// Product admin routes
		productAdmin := admin.Group("/api/products")
		{
			productAdmin.POST("", productHandler.CreateProduct)
			productAdmin.PUT("/:id", productHandler.UpdateProduct)
			productAdmin.DELETE("/:id", productHandler.DeleteProduct)
			productAdmin.POST("/:id/variants", variantHandler.CreateProductVariant)
			productAdmin.PUT("/:id/variants/:variantId", variantHandler.UpdateProductVariant)
			productAdmin.DELETE("/:id/variants/:variantId", variantHandler.DeleteProductVariant)
			productAdmin.POST("/:id/variant-combinations", variantHandler.CreateVariantCombination)
			productAdmin.PUT("/:id/variant-combinations/:combinationId", variantHandler.UpdateVariantCombination)
			productAdmin.DELETE("/:id/variant-combinations/:combinationId", variantHandler.DeleteVariantCombination)
		}

		// Direct variant admin routes
		variantAdmin := admin.Group("/api/variants")
		{
			variantAdmin.PUT("/:variantId", variantHandler.UpdateProductVariant)
			variantAdmin.DELETE("/:variantId", variantHandler.DeleteProductVariant)
		}

		variantCombinationAdmin := admin.Group("/api/variant-combinations")
		{
			variantCombinationAdmin.PUT("/:combinationId", variantHandler.UpdateVariantCombination)
			variantCombinationAdmin.DELETE("/:combinationId", variantHandler.DeleteVariantCombination)
		}

		// Order admin routes
		orderAdmin := admin.Group("/api/orders")
		{
			orderAdmin.GET("", orderHandler.ListAllOrders)
			orderAdmin.PUT("/:id/status", orderHandler.UpdateOrderStatus)
		}

		// Payment admin routes
		paymentAdmin := admin.Group("/api/payments")
		{
			paymentAdmin.GET("", paymentHandler.ListAllPayments)
			paymentAdmin.PUT("/:id/status", paymentHandler.UpdatePaymentStatus)
		}

		// Report admin routes
		reportAdmin := admin.Group("/api/reports")
		{
			reportAdmin.GET("/orders", reportHandler.GetOrdersReport)
			reportAdmin.GET("/low-stock", reportHandler.GetLowStockProductsReport)
			reportAdmin.GET("/revenue", reportHandler.GetDailyRevenueReport)
		}
	}

	// Webhook routes (no requieren autenticación pero pueden verificarse con API Key)
	webhookGroup := router.Group("/webhooks")
	{
		webhookGroup.POST("/mercadopago", paymentHandler.MercadopagoWebhook)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "el-campeon-web",
		})
	})
}
