package handler

import (
	"context"
	"logistics-platform/lib/database"
	"logistics-platform/lib/models"
	"logistics-platform/lib/token"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminLogin(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := database.PostgreSQLConn.QueryRow(
		context.Background(),
		"SELECT id, name, email FROM admin WHERE email = $1 AND password = $2",
		user.Email, user.Password,
	)

	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials", "message": err.Error()})
		return
	}

	// Generate JWT token
	token, err := token.GenerateToken(u.ID, u.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "name": u.Name})
}
