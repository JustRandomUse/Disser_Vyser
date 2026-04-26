package api

import (
	"air-quality-monitor/back/internal/handler"
	"air-quality-monitor/back/internal/service"
	"net/http"
	"os"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
	once   sync.Once
)

func initRouter() *gin.Engine {
	once.Do(func() {
		apiKey := os.Getenv("SENSOR_API_KEY")
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:5173,http://localhost:5174,http://localhost:3000"
		}

		svc := service.NewService(apiKey)
		h := handler.NewHandler(svc)

		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.Use(gin.Recovery())

		config := cors.DefaultConfig()
		config.AllowOrigins = parseOrigins(allowedOrigins)
		config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
		config.AllowCredentials = true
		r.Use(cors.New(config))

		api := r.Group("/api")
		{
			api.GET("/health", h.HealthCheck)
			api.GET("/debug/archive-check", h.DebugArchiveCheck)
			api.GET("/datasets", h.GetDataSets)
			api.GET("/datasets/:code", h.GetDataSetDetail)
			api.GET("/datasets/:code/last", h.GetLastData)
			api.GET("/datasets/:code/data", h.GetDataByDateTime)
			api.GET("/datasets/:code/aggregated", h.GetAggregatedData)
			api.GET("/datasets/:code/aggregated-stats", h.GetAggregatedDataWithStats)
			api.GET("/datasets/:code/timeseries", h.GetTimeSeries)
			api.GET("/datasets/:code/statistics", h.GetStatistics)
		}

		router = r
	})

	return router
}

func Handler(w http.ResponseWriter, r *http.Request) {
	initRouter().ServeHTTP(w, r)
}

func parseOrigins(origins string) []string {
	var result []string
	current := ""
	for i := 0; i < len(origins); i++ {
		if origins[i] == ',' {
			if current != "" {
				result = append(result, trimSpace(current))
			}
			current = ""
		} else {
			current += string(origins[i])
		}
	}
	if current != "" {
		result = append(result, trimSpace(current))
	}
	return result
}

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