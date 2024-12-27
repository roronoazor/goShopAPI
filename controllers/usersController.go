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
		Username string          `json:"username" binding:"required"`
		Email    string          `json:"email" binding:"required,email"`
		Password string          `json:"password" binding:"required"`
		Role     models.UserRole `json:"role"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid input",
			Data:    libs.NewValidationError(err),
		})
		return
	}

	// Validate role
	if err := body.Role.ValidateForSignup(); err != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid role",
			Data: []libs.ValidationError{{
				Field:   "role",
				Message: err.Error(),
			}},
		})
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

			c.JSON(http.StatusBadRequest, ProductResponse{
				Status:  "error",
				Message: "Password validation failed",
				Data:    errors,
			})
			return
		}
	}

	// Check for existing username
	var existingUser models.User
	if result := initializers.DB.Where("username = ?", body.Username).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, ProductResponse{
			Status:  "error",
			Message: "Registration failed",
			Data: []libs.ValidationError{{
				Field:   "username",
				Message: "This username is already taken",
			}},
		})
		return
	}

	// Check for existing email
	if result := initializers.DB.Where("email = ?", body.Email).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, ProductResponse{
			Status:  "error",
			Message: "Registration failed",
			Data: []libs.ValidationError{{
				Field:   "email",
				Message: "This email is already registered",
			}},
		})
		return
	}

	// Hash password and create user
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		log.Println("Failed to hash password", err)
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "An error occurred, please contact support",
		})
		return
	}

	user := models.User{
		Username: body.Username,
		Email:    body.Email,
		Password: string(hash),
		Role:     models.UserRole(body.Role),
	}

	if result := initializers.DB.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to create user",
		})
		return
	}

	tokenResponse, err := services.GenerateToken(user)
	if err != nil {
		log.Println("Failed to generate token", err)
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		Status:  "success",
		Message: "User registered successfully",
		Data:    tokenResponse,
	})
}

func Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid input",
			Data:    libs.NewValidationError(err),
		})
		return
	}

	var user models.User
	if result := initializers.DB.First(&user, "email = ?", body.Email); result.Error != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid credentials",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid credentials",
		})
		return
	}

	tokenResponse, err := services.GenerateToken(user)
	if err != nil {
		log.Println("Failed to generate token", err)
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to authenticate user",
		})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		Status:  "success",
		Message: "Login successful",
		Data:    tokenResponse,
	})
}
