package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		logger.Info(c.Request.Context(), "HTTP request",
			logging.String("method", method),
			logging.String("path", path),
			logging.Int("status", statusCode),
			logging.String("duration", duration.String()),
			logging.String("client_ip", clientIP),
		)
	}
}
