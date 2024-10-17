package service

import (
	"context"
	"logistics-platform/lib/database"
	"logistics-platform/lib/models"
	"logistics-platform/lib/token"
	"logistics-platform/lib/utils"
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

func (s *authService) RegisterUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := s.db.Exec(
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

func (s *authService) Login(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := s.db.QueryRow(
		context.Background(),
		"SELECT id, name, email FROM users WHERE email = $1 AND password = $2",
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

func (s *authService) GetProfile(c *gin.Context) {
	authUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userReq := authUser.(utils.UserRequest)

	row := s.db.QueryRow(
		context.Background(),
		"SELECT id, name, email FROM users WHERE id = $1",
		userReq.UserID,
	)

	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (s *authService) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "User service is healthy"})
}

func (s *authService) RegisterDriver(c *gin.Context) {
	var driver models.VehicleDriver
	if err := c.ShouldBindJSON(&driver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := s.db.Exec(
		context.Background(),
		"INSERT INTO vehicle_drivers (vehicle_id, name, email, password, vehicle_type, vehicle_volume) VALUES ($1, $2, $3, $4, $5, $6)",
		driver.VehicleID, driver.Name, driver.Email, driver.Password, driver.VehicleType, driver.VehicleVolume,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save driver", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver registered successfully"})
}

func (s *authService) GetDriverProfile(c *gin.Context) {
	authUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access"})
		return
	}

	user := authUser.(utils.UserRequest)

	driverID := c.Param("id")

	if user.UserID != driverID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access"})
		return
	}

	row := s.db.QueryRow(
		context.Background(),
		"SELECT id, name, email, vehicle_type, vehicle_volume FROM vehicle_drivers WHERE id = $1",
		driverID,
	)

	var driver models.VehicleDriver
	err := row.Scan(&driver.ID, &driver.Name, &driver.Email, &driver.VehicleType, &driver.VehicleVolume)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get driver", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"driver": driver})
}

func (s *authService) LoginDriver(c *gin.Context) {
	var driver models.VehicleDriver
	if err := c.ShouldBindJSON(&driver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := s.db.QueryRow(
		context.Background(),
		"SELECT id, name, email, vehicle_type, vehicle_volume FROM vehicle_drivers WHERE email = $1 AND password = $2",
		driver.Email, driver.Password,
	)

	err := row.Scan(&driver.ID, &driver.Name, &driver.Email, &driver.VehicleType, &driver.VehicleVolume)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login driver", "message": err.Error()})
		return
	}

	token, err := token.GenerateToken(driver.ID, driver.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "ID": driver.ID, "name": driver.Name})
}

func (s *authService) AdminLogin(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := s.db.QueryRow(
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
