package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"heart-rate-server/internal/models"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UUIDCacheMiddleware struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewUUIDCacheMiddleware(db *gorm.DB, redis *redis.Client) *UUIDCacheMiddleware {
	return &UUIDCacheMiddleware{
		DB:    db,
		Redis: redis,
	}
}

// Handler 缓存UUID到UserID的映射
func (m *UUIDCacheMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 只处理包含UUID的路径
		vars := mux.Vars(r)
		uuid, ok := vars["uuid"]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		cacheKey := fmt.Sprintf("uuid_to_user_id:%s", uuid)

		// 1. 尝试从缓存获取
		var userID uint
		if cachedID, err := m.Redis.Get(ctx, cacheKey).Uint64(); err == nil {
			userID = uint(cachedID)
			ctx = context.WithValue(ctx, "cached_user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 2. 缓存未命中，查询数据库
		var user struct {
			ID uint
		}
		if err := m.DB.Model(&models.User{}).Select("id").Where("uuid = ?", uuid).First(&user).Error; err != nil {
			// 缓存空结果防止穿透
			if errors.Is(err, gorm.ErrRecordNotFound) {
				m.Redis.Set(ctx, cacheKey, "null", 5*time.Minute)
			}
			next.ServeHTTP(w, r)
			return
		}

		// 3. 写入缓存
		userID = user.ID
		m.Redis.Set(ctx, cacheKey, userID, time.Hour)
		ctx = context.WithValue(ctx, "cached_user_id", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
