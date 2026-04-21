package main

import (
	"air-quality-monitor/back/internal/handler"
	"air-quality-monitor/back/internal/service"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("SENSOR_API_KEY")
	if apiKey == "" {
		log.Fatal("SENSOR_API_KEY environment variable is required")
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize service
	svc := service.NewService(apiKey)

	// Initialize handler
	h := handler.NewHandler(svc)

	// Setup Gin router
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	router.Use(cors.New(config))

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", h.HealthCheck)

		// Datasets
		api.GET("/datasets", h.GetDataSets)
		api.GET("/datasets/:code", h.GetDataSetDetail)
		api.GET("/datasets/:code/last", h.GetLastData)
		api.GET("/datasets/:code/data", h.GetDataByDateTime)
		api.GET("/datasets/:code/aggregated", h.GetAggregatedData)
		api.GET("/datasets/:code/aggregated-stats", h.GetAggregatedDataWithStats)
		api.GET("/datasets/:code/timeseries", h.GetTimeSeries)
		api.GET("/datasets/:code/statistics", h.GetStatistics)
	}

	// Serve frontend static files
	router.Static("/assets", "../front/dist/assets")
	router.StaticFile("/", "../front/dist/index.html")
	router.NoRoute(func(c *gin.Context) {
		c.File("../front/dist/index.html")
	})

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
