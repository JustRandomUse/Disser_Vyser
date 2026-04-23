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

	// Get allowed origins from environment
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		// Default for local development
		allowedOrigins = "http://localhost:5173,http://localhost:5174,http://localhost:3000"
	}

	// Initialize service
	svc := service.NewService(apiKey)

	// Initialize handler
	h := handler.NewHandler(svc)

	// Setup Gin router
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = parseOrigins(allowedOrigins)
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	config.AllowCredentials = true
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

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// parseOrigins splits comma-separated origins string into slice
func parseOrigins(origins string) []string {
	var result []string
	for _, origin := range splitAndTrim(origins, ",") {
		if origin != "" {
			result = append(result, origin)
		}
	}
	return result
}

// splitAndTrim splits string by separator and trims whitespace
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		parts = append(parts, trimmed)
	}
	return parts
}

// splitString splits string by separator
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
