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

func NewAdminService(redisClient *redis.Client, pool *pgxpool.Pool, cache *Cache) interfaces.AdminInterface {
	return &AdminService{redisClient: redisClient, pool: pool, cache: cache}
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]cacheItem),
	}
}
