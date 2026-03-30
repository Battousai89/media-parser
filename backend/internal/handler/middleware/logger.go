package middleware

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type LoggerMiddleware struct{}

func NewLoggerMiddleware() *LoggerMiddleware {
	return &LoggerMiddleware{}
}

func (m *LoggerMiddleware) Logger() gin.HandlerFunc {
	logger := log.New(os.Stdout, "[GIN] ", log.LstdFlags)

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		logger.Printf("%s %d %s %s %s %v\n",
			method,
			statusCode,
			path,
			query,
			clientIP,
			latency,
		)

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Printf("Error: %v\n", e)
			}
		}
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Fprintf(os.Stderr, "[PANIC] %v\n", err)
				c.AbortWithStatus(500)
			}
		}()

		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
