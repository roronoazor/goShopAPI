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
	{
		products.GET("/", controllers.GetProducts)
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
