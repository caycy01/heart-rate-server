package storage

import (
	"context"
	"fmt"
	"heart-rate-server/internal/config"

	"time"

	"github.com/go-redis/redis/v8"
)

func InitRedis(cfg *config.Config) (*redis.Client, error) {
	// 打印账号密码
	fmt.Printf("redis addr: %s, password: %s, db: %d\n", cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return client, nil
}
