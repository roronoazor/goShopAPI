package middlewares

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/roronoazor/goShopAPI/initializers"
	"github.com/roronoazor/goShopAPI/models"
)

func RequireAuth(c *gin.Context) {
	log.Println("RequireAuth middleware")

	tokenString := c.GetHeader("Authorization")

	// Remove "Bearer " prefix if it exists
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		c.Abort()
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		log.Println("Error parsing token:", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Failed to authenticate",
		})
		c.Abort()
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		// check if token is expired
		if time.Now().Unix() > int64(claims["exp"].(float64)) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token expired",
			})
			c.Abort()
			return
		}

		// get user id from token
		var user models.User
		initializers.DB.First(&user, claims["sub"])

		if user.ID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// attach user to request
		c.Set("user", user)

		// pass to next middleware
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
