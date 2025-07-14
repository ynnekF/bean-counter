// main.go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ynnekF/bean-counter/internal" // Adjust the import path as necessary
	"github.com/ynnekF/bean-counter/middleware"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger = internal.CreateLogger()

// getRouter initializes the Gin router with middleware and returns it.
func getRouter() *gin.Engine {
	router := gin.New()

	// Gin's built-in request logger and panic recovery middleware
	router.Use(gin.Recovery())

	// Custom middleware
	router.Use(middleware.RequestLoggerMiddleware(logger))

	return router
}

// startApi registers the API routes and starts the HTTP server.
func startApi(router *gin.Engine) {
	logger.Info("Starting API server...")
	router.Static("/app", "./static")

	// General API routes.
	router.GET("/api/current-day", internal.GetCurrentDay)
	router.GET("/api/who-pays-next", internal.WhoPays)
	router.GET("/api/balances", internal.GetBalances)
	router.GET("/api/ledger", internal.GetLedger)

	// User management routes.
	router.POST("/api/update-user-field", internal.UpdateUserField)
	router.DELETE("/api/delete-user", internal.DeleteUser)
	router.POST("/api/reset-user", internal.ResetUser)
	router.POST("/api/add-user", internal.CreateUser)

	// Coffee run management routes.
	router.POST("/api/coffee-run", internal.RecordRun)

	// Start the server on port 8080 (blocking call)
	if err := router.Run(":8080"); err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}

}

func main() {
	logger.Info("Starting Bean Counter server...")

	gin.SetMode(gin.ReleaseMode)

	router := getRouter()
	startApi(router)
}
