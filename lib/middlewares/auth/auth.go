package auth

import (
	"logistics-platform/lib/token"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenStr := strings.Split(authHeader, "Bearer ")[1]
		valid, err := token.ValidateToken(tokenStr)
		if err != nil || !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Next()
	}
}

func AuthInjectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		headerAuthToken := c.GetHeader("Authorization")
		if headerAuthToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
			return
		}

		authToken := strings.TrimPrefix(headerAuthToken, "Bearer ")
		if authToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
			return
		}

		// validate token
		user, err := token.GetUserFromToken(authToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
