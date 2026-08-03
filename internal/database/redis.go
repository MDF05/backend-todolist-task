package database

import (
	"context"
	"log"
	"todo-api/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates and returns a new Redis client
func NewRedisClient(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Ping to verify connection
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Printf("Warning: Redis connection failed: %v — caching will be disabled", err)
		return nil
	}

	log.Println("Redis connected successfully")
	return client
}
