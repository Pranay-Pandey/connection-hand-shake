package interfaces

import (
	"context"
	"logistics-platform/lib/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PricingInterface interface {
	HandlePriceEstimate(c *gin.Context)
	EstimatePrice(ctx context.Context, req models.BookingRequest) (models.PriceEstimate, error)
	GetVehiclePricing(vehicleType string) (models.VehiclePricing, error)
	CalculateSurgeMultiplier(ctx context.Context, pickup, dropoff models.GeoPoint) float64
	GetCurrentDemand(location models.GeoPoint) (float64, error)
	GracefulShutdown(server *http.Server)
}
