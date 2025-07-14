package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLoggerMiddleware logs incoming requests and any errors that occur during processing.
func RequestLoggerMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	logger.Info("RequestLoggerMiddleware initialized")

	return func(c *gin.Context) {
		start := time.Now()

		// Log basic request info at the beginning of the request lifecycle
		logger.Infow("incoming request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)

		// Process the request - calling next handler in the chain
		c.Next()

		// Compute latency
		latency := time.Since(start)

		// Log basic request/response details
		logger.Infow("request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", latency,
			"client_ip", c.ClientIP(),
		)

		// Log any errors collected during request handling
		for _, err := range c.Errors {
			logger.Errorw("error occurred while handling request",
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)
		}
	}
}
