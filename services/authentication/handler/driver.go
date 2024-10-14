package handler

import (
	"context"
	"log"
	"logistics-platform/lib/database"
	"logistics-platform/lib/models"
	"logistics-platform/lib/token"
	"logistics-platform/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterDriver(c *gin.Context) {
	var driver models.VehicleDriver
	if err := c.ShouldBindJSON(&driver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.PostgreSQLConn.Exec(
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

func GetDriverProfile(c *gin.Context) {
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

	row := database.PostgreSQLConn.QueryRow(
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

func LoginDriver(c *gin.Context) {
	var driver models.VehicleDriver
	if err := c.ShouldBindJSON(&driver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := database.PostgreSQLConn.QueryRow(
		context.Background(),
		"SELECT id, name, email, vehicle_type, vehicle_volume FROM vehicle_drivers WHERE email = $1 AND password = $2",
		driver.Email, driver.Password,
	)

	err := row.Scan(&driver.ID, &driver.Name, &driver.Email, &driver.VehicleType, &driver.VehicleVolume)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login driver", "message": err.Error()})
		return
	}

	log.Println("driver", driver)
	token, err := token.GenerateToken(driver.ID, driver.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "ID": driver.ID, "name": driver.Name})
}
