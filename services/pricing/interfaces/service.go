package interfaces

import (
	"context"
	"logistics-platform/lib/utils"
	"logistics-platform/services/pricing/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PricingInterface interface {
	HandlePriceEstimate(c *gin.Context)
	EstimatePrice(ctx context.Context, req utils.BookingRequest) (models.PriceEstimate, error)
	GetVehiclePricing(vehicleType string) (models.VehiclePricing, error)
	CalculateSurgeMultiplier(ctx context.Context, pickup, dropoff utils.GeoPoint) float64
	GetCurrentDemand(location utils.GeoPoint) (float64, error)
	GracefulShutdown(server *http.Server)
}
