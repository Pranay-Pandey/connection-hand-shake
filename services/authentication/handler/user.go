package handler

import (
	"context"
	"logistics-platform/lib/database"
	"logistics-platform/lib/models"
	"logistics-platform/lib/token"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.PostgreSQLConn.Exec(
		context.Background(),
		"INSERT INTO users (name, email, password) VALUES ($1, $2, $3)",
		user.Name, user.Email, user.Password,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func Login(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := database.PostgreSQLConn.QueryRow(
		context.Background(),
		"SELECT id, name, email FROM users WHERE email = $1 AND password = $2",
		user.Email, user.Password,
	)

	println(user.Email, user.Password)

	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials", "message": err.Error()})
		return
	}

	// Generate JWT token
	token, err := token.GenerateToken(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func GetProfile(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	tokenStr := authHeader[7:]

	userID, err := token.GetSubjectFromToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	row := database.PostgreSQLConn.QueryRow(
		context.Background(),
		"SELECT id, name, email FROM users WHERE id = $1",
		userID,
	)

	var user models.User
	err = row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "User service is healthy"})
}
