package services

import (
	"os"
	"time"

	"github.com/roronoazor/goShopAPI/models"

	"github.com/golang-jwt/jwt/v4"
)

type tokenResponse struct {
	Token    string `json:"token"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func GenerateToken(user models.User) (tokenResponse, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		return tokenResponse{}, err
	}

	return tokenResponse{
		Token:    tokenString,
		Email:    user.Email,
		Username: user.Username,
	}, nil
}
