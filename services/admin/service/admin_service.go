package service

import (
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"logistics-platform/services/admin/interfaces"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

type AdminService struct {
	redisClient *redis.Client
	pool        *pgxpool.Pool
	cache       *Cache
}

type FleetStats struct {
	TotalVehicles        int            `json:"totalVehicles"`
	ActiveVehicles       int            `json:"activeVehicles"`
	VehicleTypeBreakdown map[string]int `json:"vehicleTypeBreakdown"`
}

type DriverPerformance struct {
	DriverID     int     `json:"driverID"`
	Name         string  `json:"name"`
	TripCount    int     `json:"tripCount"`
	AvgTripTime  float64 `json:"avgTripTime"`
	TotalRevenue float64 `json:"totalRevenue"`
}

type BookingAnalytics struct {
	TotalBookings     int     `json:"totalBookings"`
	CompletedBookings int     `json:"completedBookings"`
	CancelledBookings int     `json:"cancelledBookings"`
	AvgTripTime       float64 `json:"avgTripTime"`
	TotalRevenue      float64 `json:"totalRevenue"`
}

func NewAdminService(redisClient *redis.Client, pool *pgxpool.Pool, cache *Cache) interfaces.AdminInterface {
	return &AdminService{redisClient: redisClient, pool: pool, cache: cache}
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]cacheItem),
	}
}
