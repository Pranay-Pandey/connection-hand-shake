package token

import (
	"fmt"
	"strconv"
	"time"

	"logistics-platform/lib/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

var jwtKey = []byte(viper.GetString("JWT_SECRET"))

// GenerateToken generates a JWT token for the user with userID and userName
func GenerateToken(userID int32, userName string) (string, error) {
	// Convert userID from int32 to string
	userIDStr := strconv.Itoa(int(userID))

	expirationTime := time.Now().Add(72 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   userIDStr,
		Audience:  userName,
		Issuer:    "logistics-platform",
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	return tokenString, nil
}

// ValidateToken validates the provided JWT token
func ValidateToken(tokenString string) (bool, error) {
	claims := &jwt.StandardClaims{}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	// Check for errors or invalid tokens
	if err != nil {
		return false, fmt.Errorf("error parsing token: %v", err)
	}

	return token.Valid, nil
}

// GetUserFromToken extracts user information from the token
func GetUserFromToken(tokenString string) (utils.UserRequest, error) {
	claims := &jwt.StandardClaims{}

	// Parse the token
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return utils.UserRequest{}, fmt.Errorf("error parsing token: %v", err)
	}

	user := utils.UserRequest{
		UserID:   claims.Subject,
		UserName: claims.Audience,
	}

	return user, nil
}
