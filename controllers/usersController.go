package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roronoazor/goShopAPI/initializers"
	"github.com/roronoazor/goShopAPI/libs"
	"github.com/roronoazor/goShopAPI/models"
	"github.com/roronoazor/goShopAPI/services"
	"github.com/roronoazor/goShopAPI/validators"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {

	var body struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, libs.NewValidationError(err))
		return
	}

	// Validate password complexity
	if err := validators.ValidatePassword(body.Password); err != nil {
		if pErr, ok := err.(validators.PasswordError); ok {
			var errors []libs.ValidationError

			if pErr.MinLength {
				errors = append(errors, libs.ValidationError{
					Field:   "password",
					Message: "Password must be at least 8 characters long",
				})
			}
			if pErr.UpperCase {
				errors = append(errors, libs.ValidationError{
					Field:   "password",
					Message: "Password must contain at least one uppercase letter",
				})
			}
			if pErr.LowerCase {
				errors = append(errors, libs.ValidationError{
					Field:   "password",
					Message: "Password must contain at least one lowercase letter",
				})
			}
			if pErr.Number {
				errors = append(errors, libs.ValidationError{
					Field:   "password",
					Message: "Password must contain at least one number",
				})
			}
			if pErr.SpecialChar {
				errors = append(errors, libs.ValidationError{
					Field:   "password",
					Message: "Password must contain at least one special character",
				})
			}

			c.JSON(http.StatusBadRequest, libs.ErrorResponse{
				Status:  "error",
				Message: "Password validation failed",
				Errors:  errors,
			})
			return
		}
	}

	// Check for existing username
	var existingUser models.User
	if result := initializers.DB.Where("username = ?", body.Username).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, libs.ErrorResponse{
			Status:  "error",
			Message: "Registration failed",
			Errors: []libs.ValidationError{
				{
					Field:   "username",
					Message: "This username is already taken",
				},
			},
		})
		return
	}

	// Check for existing email
	if result := initializers.DB.Where("email = ?", body.Email).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, libs.ErrorResponse{
			Status:  "error",
			Message: "Registration failed",
			Errors: []libs.ValidationError{
				{
					Field:   "email",
					Message: "This email is already registered",
				},
			},
		})
		return
	}

	// hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		// log the error
		log.Println("Failed to hash password", err)

		// return an error message
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "An error occurred, please contact support",
		})
		return
	}

	user := models.User{
		Username: body.Username,
		Email:    body.Email,
		Password: string(hash),
	}
	result := initializers.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "An error occurred, please contact support",
		})
		return
	}

	tokenResponse, err := services.GenerateToken(user)

	if err != nil {

		log.Println("Failed to generate token", err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user, please contact support",
		})
		return
	}

	c.JSON(http.StatusOK, tokenResponse)

}

func Login(c *gin.Context) {
	// Get email and password from body
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, libs.NewValidationError(err))
		return
	}

	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, libs.ErrorResponse{
			Status:  "error",
			Message: "Invalid Credentials",
		})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		c.JSON(http.StatusBadRequest, libs.ErrorResponse{
			Status:  "error",
			Message: "Invalid Credentials",
		})
		return
	}

	// generate and return a jwt token
	tokenResponse, err := services.GenerateToken(user)

	if err != nil {

		log.Println("Failed to generate token", err)

		c.JSON(http.StatusInternalServerError, libs.ErrorResponse{
			Status:  "error",
			Message: "Failed to authenticate user",
		})
		return
	}

	c.JSON(http.StatusOK, tokenResponse)

}
