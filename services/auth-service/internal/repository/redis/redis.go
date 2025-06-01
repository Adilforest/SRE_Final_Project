package redisrepo

import (
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASS")
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})
}
