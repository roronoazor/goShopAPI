package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roronoazor/goShopAPI/controllers"
	"github.com/roronoazor/goShopAPI/initializers"
	"github.com/roronoazor/goShopAPI/middlewares"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDb()
}

func main() {
	r := gin.Default()

	// auth routes under /auth
	auth := r.Group("/auth")
	{
		auth.POST("/signup", controllers.SignUp)
		auth.POST("/login", controllers.Login)
	}

	// products routes under /products
	products := r.Group("/products")
	products.Use(middlewares.RequireAuth)
	//products.Use(middlewares.RequireAdmin()) // only admin can access these routes
	{
		products.POST("/", controllers.CreateProduct)
		products.GET("/", controllers.GetProducts)
		products.GET("/:id", controllers.GetProduct)
		products.PUT("/:id", controllers.UpdateProduct)
		products.DELETE("/:id", controllers.DeleteProduct)
	}

	// Order routes
	orders := r.Group("/orders")
	orders.Use(middlewares.RequireAuth)
	{
		orders.POST("/", controllers.CreateOrder)
		orders.GET("/", controllers.GetUserOrders)
		orders.GET("/:id", controllers.GetOrder) // Add this line
		orders.POST("/:id/cancel", controllers.CancelOrder)

		// Admin only routes
		admin := orders.Group("/")
		admin.Use(middlewares.RequireAdmin())
		{
			admin.PUT("/:id/status", controllers.UpdateOrderStatus)
		}
	}

	// Custom 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"status":    "error",
			"message":   "Route not found",
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	r.Run()
}
