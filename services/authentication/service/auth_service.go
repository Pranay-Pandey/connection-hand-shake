package service

import (
	"logistics-platform/lib/database"
	"logistics-platform/services/authentication/interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

type authService struct {
	db database.Database
}

func NewAuthService(db database.Database) interfaces.AuthService {
	return &authService{db: db}
}

func (s *authService) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "User service is healthy"})
}
