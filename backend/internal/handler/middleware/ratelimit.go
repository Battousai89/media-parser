package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/media-parser/backend/internal/model/dto"
	"golang.org/x/time/rate"
)

type RateLimitMiddleware struct {
	limiter *rate.Limiter
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

func NewRateLimitMiddleware(cfg RateLimitConfig) *RateLimitMiddleware {
	if cfg.RequestsPerSecond <= 0 {
		cfg.RequestsPerSecond = 10
	}
	if cfg.BurstSize <= 0 {
		cfg.BurstSize = 20
	}

	return &RateLimitMiddleware{
		limiter: rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.BurstSize),
	}
}

func (m *RateLimitMiddleware) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.limiter.Allow() {
			c.JSON(429, dto.Response{
				Success: false,
				Error:   &dto.ErrorData{Code: "RATE_LIMITED", Message: "Too many requests"},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

type PerKeyRateLimitMiddleware struct {
	limiters map[string]*rate.Limiter
	config   RateLimitConfig
}

func NewPerKeyRateLimitMiddleware(cfg RateLimitConfig) *PerKeyRateLimitMiddleware {
	return &PerKeyRateLimitMiddleware{
		limiters: make(map[string]*rate.Limiter),
		config:   cfg,
	}
}

func (m *PerKeyRateLimitMiddleware) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if tokenID, exists := c.Get("token_id"); exists {
			switch v := tokenID.(type) {
			case int:
				key = fmt.Sprintf("token:%d", v)
			case string:
				key = v
			default:
				key = c.ClientIP()
			}
		}

		limiter := m.getLimiter(key)
		if !limiter.Allow() {
			c.JSON(429, dto.Response{
				Success: false,
				Error:   &dto.ErrorData{Code: "RATE_LIMITED", Message: "Too many requests"},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (m *PerKeyRateLimitMiddleware) getLimiter(key string) *rate.Limiter {
	limiter, exists := m.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(m.config.RequestsPerSecond), m.config.BurstSize)
		m.limiters[key] = limiter
	}
	return limiter
}

func (m *PerKeyRateLimitMiddleware) Cleanup() {
	// Простая реализация - можно добавить TTL для каждого limiter
}
