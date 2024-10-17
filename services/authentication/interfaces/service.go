package interfaces

import (
	"github.com/gin-gonic/gin"
)

type AuthService interface {
	RegisterUser(c *gin.Context)
	Login(c *gin.Context)
	GetProfile(c *gin.Context)
	HealthCheck(c *gin.Context)
	RegisterDriver(c *gin.Context)
	LoginDriver(c *gin.Context)
	GetDriverProfile(c *gin.Context)
	AdminLogin(c *gin.Context)
}
