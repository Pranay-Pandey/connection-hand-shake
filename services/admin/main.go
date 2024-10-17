package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"logistics-platform/lib/config"
	"logistics-platform/lib/database"
	"logistics-platform/lib/middlewares/cors"
	"logistics-platform/lib/models"
	"logistics-platform/lib/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/redis/go-redis/v9"
)

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

type Cache struct {
	sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]cacheItem),
	}
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.Lock()
	defer c.Unlock()
	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(duration),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()
	item, found := c.items[key]
	if !found {
		return nil, false
	}
	if time.Now().After(item.expiration) {
		delete(c.items, key)
		return nil, false
	}
	return item.value, true
}

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to parse pool config: %v", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	redisClient, err := database.InitRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	service := &AdminService{
		pool:        pool,
		redisClient: redisClient,
		cache:       NewCache(),
	}

	router := gin.Default()
	router.Use(cors.CORSMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	adminGroup := router.Group("/admin")
	adminGroup.GET("/fleet-stats", service.getFleetStats)
	adminGroup.GET("/driver-performance", service.getDriverPerformance)
	adminGroup.GET("/booking-analytics", service.getBookingAnalytics)
	adminGroup.GET("/vehicle-locations", service.getVehicleLocations)
	adminGroup.POST("/update-vehicle", service.updateVehicle)

	server := &http.Server{
		Addr:    ":8085",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	utils.WaitForShutdown(server)
}

func (s *AdminService) getFleetStats(c *gin.Context) {
	if stats, found := s.cache.Get("fleet_stats"); found {
		c.JSON(http.StatusOK, stats)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stats FleetStats

	err := retry(3, 100*time.Millisecond, func() error {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM vehicle_drivers`).Scan(&stats.TotalVehicles)
		if err != nil {
			return fmt.Errorf("failed to fetch total vehicles: %v", err)
		}

		err = tx.QueryRow(ctx, `SELECT COUNT(DISTINCT driver_id) FROM booking WHERE status NOT IN ('cancelled', 'completed')`).Scan(&stats.ActiveVehicles)
		if err != nil {
			return fmt.Errorf("failed to fetch active vehicles: %v", err)
		}

		rows, err := tx.Query(ctx, `SELECT vehicle_type, COUNT(*) FROM vehicle_drivers GROUP BY vehicle_type`)
		if err != nil {
			return fmt.Errorf("failed to fetch vehicle type breakdown: %v", err)
		}
		defer rows.Close()

		stats.VehicleTypeBreakdown = make(map[string]int)
		for rows.Next() {
			var vehicleType string
			var count int
			if err := rows.Scan(&vehicleType, &count); err != nil {
				return fmt.Errorf("failed to scan vehicle type breakdown: %v", err)
			}
			stats.VehicleTypeBreakdown[vehicleType] = count
		}

		return tx.Commit(ctx)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.cache.Set("fleet_stats", stats, 5*time.Minute)
	c.JSON(http.StatusOK, stats)
}

func (s *AdminService) getDriverPerformance(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var performances []DriverPerformance

	err := retry(3, 100*time.Millisecond, func() error {
		rows, err := s.pool.Query(ctx, `
			SELECT 
				vd.id AS driver_id,
				vd.name,
				COUNT(b.id) AS trip_count,
				AVG(EXTRACT(EPOCH FROM (b.completed_at - b.created_at))) AS avg_trip_time,
				SUM(CAST(b.price AS FLOAT)) AS total_revenue
			FROM 
				vehicle_drivers vd
			LEFT JOIN 
				booking b ON vd.id = b.driver_id AND b.status = 'completed'
			WHERE
				b.completed_at IS NOT NULL
			GROUP BY 
				vd.id, vd.name
			ORDER BY 
				trip_count DESC
		`)
		if err != nil {
			return fmt.Errorf("failed to fetch driver performance: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var perf DriverPerformance
			if err := rows.Scan(&perf.DriverID, &perf.Name, &perf.TripCount, &perf.AvgTripTime, &perf.TotalRevenue); err != nil {
				return fmt.Errorf("failed to scan driver performance: %v", err)
			}
			perf.AvgTripTime = perf.AvgTripTime / 60
			performances = append(performances, perf)
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, performances)
}

func (s *AdminService) getBookingAnalytics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var analytics BookingAnalytics

	err := retry(3, 100*time.Millisecond, func() error {
		return s.pool.QueryRow(ctx, `
			SELECT 
				COUNT(*) AS total_bookings,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed_bookings,
				COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_bookings,
				AVG(EXTRACT(EPOCH FROM (completed_at - created_at))) FILTER (WHERE status = 'completed') AS avg_trip_time,
				SUM(CAST(price AS FLOAT)) FILTER (WHERE status = 'completed') AS total_revenue
			FROM booking
			WHERE
				completed_at IS NOT NULL
		`).Scan(
			&analytics.TotalBookings,
			&analytics.CompletedBookings,
			&analytics.CancelledBookings,
			&analytics.AvgTripTime,
			&analytics.TotalRevenue,
		)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch booking analytics: %v", err)})
		return
	}

	analytics.AvgTripTime = analytics.AvgTripTime / 60
	c.JSON(http.StatusOK, analytics)
}

func (s *AdminService) getVehicleLocations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type VehicleLocation struct {
		ID          int32   `json:"id"`
		Name        string  `json:"name"`
		VehicleType string  `json:"vehicleType"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Status      string  `json:"status"`
	}

	var locations []VehicleLocation

	err := retry(3, 100*time.Millisecond, func() error {
		rows, err := s.pool.Query(ctx, `
			SELECT 
				vd.id, 
				vd.name, 
				vd.vehicle_type,
				b.pickup_latitude, 
				b.pickup_longitude,
				b.dropoff_latitude,
				b.dropoff_longitude,
				b.status
			FROM 
				vehicle_drivers vd
			LEFT JOIN 
				booking b ON vd.id = b.driver_id AND b.status IN ('in_progress', 'enroute_to_pickup')
		`)
		if err != nil {
			return fmt.Errorf("failed to fetch vehicle locations: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var loc VehicleLocation
			var pickupLat, pickupLon, dropoffLat, dropoffLon pgtype.Float8
			var status pgtype.Text
			if err := rows.Scan(&loc.ID, &loc.Name, &loc.VehicleType, &pickupLat, &pickupLon, &dropoffLat, &dropoffLon, &status); err != nil {
				return fmt.Errorf("failed to scan vehicle location: %v", err)
			}
			if status.Status == pgtype.Present {
				loc.Status = status.String
				if status.String == "completed" || status.String == "cancelled" {
					loc.Latitude = dropoffLat.Float
					loc.Longitude = dropoffLon.Float
				} else {
					loc.Latitude = pickupLat.Float
					loc.Longitude = pickupLon.Float
				}
			} else {
				loc.Status = "idle"
			}
			// Get driver location
			driverID := fmt.Sprintf("%d", loc.ID)
			driverPos, err := s.redisClient.GeoPos(context.Background(), "driver_locations", driverID).Result()
			if err == nil && len(driverPos) > 0 {
				if driverPos != nil && driverPos[0] != nil {
					loc.Latitude = driverPos[0].Latitude
					loc.Longitude = driverPos[0].Longitude
				}
			}

			if loc.Latitude == 0 && loc.Longitude == 0 {
				loc.Status = "offline"
			}
			locations = append(locations, loc)
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, locations)
}

func (s *AdminService) updateVehicle(c *gin.Context) {
	var vehicle models.VehicleDriver
	if err := c.ShouldBindJSON(&vehicle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := retry(3, 100*time.Millisecond, func() error {
		_, err := s.pool.Exec(ctx, `
			UPDATE vehicle_drivers
			SET name = $1, vehicle_id = $2, email = $3, vehicle_type = $4, vehicle_volume = $5
			WHERE id = $6
		`, vehicle.Name, vehicle.VehicleID, vehicle.Email, vehicle.VehicleType, vehicle.VehicleVolume, vehicle.ID)
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update vehicle: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vehicle updated successfully"})
}

func retry(attempts int, sleep time.Duration, f func() error) error {
	var err error
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return nil
		}
		if i >= attempts-1 {
			break
		}
		time.Sleep(sleep)
		sleep *= 2 // exponential back-off
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
