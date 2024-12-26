package middlewares

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roronoazor/goShopAPI/models"
)

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user from the context (set by RequireAuth middleware)
		user, exists := c.Get("user")
		if !exists {
			log.Println("User not found in context")

			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// Type assert to get the user model
		u, ok := user.(models.User)
		if !ok {

			log.Println("Failed to get user from context")

			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Internal server error",
			})
			c.Abort()
			return
		}

		// Check if user is admin
		if u.Role != models.UserRoleAdmin {
			log.Println("User is not admin")

			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "Admin privileges required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
