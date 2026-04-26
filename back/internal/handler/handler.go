package handler

import (
	"air-quality-monitor/back/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests
type Handler struct {
	service *service.Service
}

// NewHandler creates a new handler
func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

// Response represents API response
type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// GetDataSets handles GET /api/datasets
func (h *Handler) GetDataSets(c *gin.Context) {
	data, err := h.service.GetDataSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// GetDataSetDetail handles GET /api/datasets/:code
func (h *Handler) GetDataSetDetail(c *gin.Context) {
	code := c.Param("code")

	data, err := h.service.GetDataSetDetail(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// GetLastData handles GET /api/datasets/:code/last
func (h *Handler) GetLastData(c *gin.Context) {
	code := c.Param("code")

	// Parse query parameters
	sites := parseIntArray(c.Query("sites"))
	indicators := parseStringArray(c.Query("indicators"))

	data, err := h.service.GetLastData(code, sites, indicators)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// GetDataByDateTime handles GET /api/datasets/:code/data
func (h *Handler) GetDataByDateTime(c *gin.Context) {
	code := c.Param("code")
	dateStr := c.Query("date")
	hourStr := c.Query("hour")

	if dateStr == "" || hourStr == "" {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  "date and hour parameters are required",
		})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  "invalid date format, use YYYY-MM-DD",
		})
		return
	}

	hour, err := strconv.Atoi(hourStr)
	if err != nil || hour < 0 || hour > 23 {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  "invalid hour, must be 0-23",
		})
		return
	}

	timeBegin := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, time.UTC)
	timeEnd := timeBegin.Add(time.Hour)

	sites := parseIntArray(c.Query("sites"))
	indicators := parseStringArray(c.Query("indicators"))

	// Check if this is current hour (use /data/last for live data)
	now := time.Now().UTC()
	currentHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)

	var data interface{}

	if timeBegin.Equal(currentHour) {
		// Use live data endpoint for current hour
		data, err = h.service.GetLastData(code, sites, indicators)
	} else {
		// Use archive aggregated data for historical hours
		// According to API.md, hourly archive data is available for last 30 days
		data, err = h.service.GetAggregatedData(code, timeBegin, timeEnd, "hour", sites, indicators)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// GetAggregatedData handles GET /api/datasets/:code/aggregated
func (h *Handler) GetAggregatedData(c *gin.Context) {
	code := c.Param("code")

	// Parse time parameters
	timeBeginStr := c.Query("time_begin")
	timeEndStr := c.Query("time_end")
	interval := c.DefaultQuery("interval", "hour")

	// Validate interval according to API.md
	// Only hour, day, month are supported by Sensor Hub API
	if interval != "hour" && interval != "day" && interval != "month" {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  "invalid interval, must be one of: hour, day, month",
		})
		return
	}

	timeBegin, err := time.Parse("2006-01-02T15:04:05Z", timeBeginStr)
	if err != nil {
		timeBegin, err = time.Parse("2006-01-02", timeBeginStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Status: "error",
				Error:  "invalid time_begin format",
			})
			return
		}
	}

	timeEnd, err := time.Parse("2006-01-02T15:04:05Z", timeEndStr)
	if err != nil {
		timeEnd, err = time.Parse("2006-01-02", timeEndStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Status: "error",
				Error:  "invalid time_end format",
			})
			return
		}
	}

	sites := parseIntArray(c.Query("sites"))
	indicators := parseStringArray(c.Query("indicators"))

	data, err := h.service.GetAggregatedData(code, timeBegin, timeEnd, interval, sites, indicators)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// GetAggregatedDataWithStats handles GET /api/datasets/:code/aggregated-stats
func (h *Handler) GetAggregatedDataWithStats(c *gin.Context) {
	code := c.Param("code")

	timeBeginStr := c.Query("time_begin")
	timeEndStr := c.Query("time_end")
	interval := c.DefaultQuery("interval", "hour")

	timeBegin, err := time.Parse("2006-01-02T15:04:05Z", timeBeginStr)
	if err != nil {
		timeBegin, err = time.Parse("2006-01-02", timeBeginStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Status: "error",
				Error:  "invalid time_begin format",
			})
			return
		}
	}

	timeEnd, err := time.Parse("2006-01-02T15:04:05Z", timeEndStr)
	if err != nil {
		timeEnd, err = time.Parse("2006-01-02", timeEndStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Status: "error",
				Error:  "invalid time_end format",
			})
			return
		}
	}

	sites := parseIntArray(c.Query("sites"))
	indicators := parseStringArray(c.Query("indicators"))

	data, err := h.service.GetAggregatedDataWithStats(code, timeBegin, timeEnd, interval, sites, indicators)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// GetTimeSeries handles GET /api/datasets/:code/timeseries
func (h *Handler) GetTimeSeries(c *gin.Context) {
	code := c.Param("code")

	timeBeginStr := c.Query("time_begin")
	timeEndStr := c.Query("time_end")
	interval := c.Query("interval")

	timeBegin, err := time.Parse("2006-01-02T15:04:05Z", timeBeginStr)
	if err != nil {
		timeBegin, err = time.Parse("2006-01-02", timeBeginStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Status: "error",
				Error:  "invalid time_begin format",
			})
			return
		}
	}

	timeEnd, err := time.Parse("2006-01-02T15:04:05Z", timeEndStr)
	if err != nil {
		timeEnd, err = time.Parse("2006-01-02", timeEndStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Status: "error",
				Error:  "invalid time_end format",
			})
			return
		}
	}

	sites := parseIntArray(c.Query("sites"))
	indicators := parseStringArray(c.Query("indicators"))

	data, err := h.service.GetTimeSeriesData(code, timeBegin, timeEnd, interval, sites, indicators)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// GetStatistics handles GET /api/datasets/:code/statistics
func (h *Handler) GetStatistics(c *gin.Context) {
	code := c.Param("code")

	timeBeginStr := c.Query("time_begin")
	timeEndStr := c.Query("time_end")

	timeBegin, err := time.Parse("2006-01-02T15:04:05Z", timeBeginStr)
	if err != nil {
		timeBegin, err = time.Parse("2006-01-02", timeBeginStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Status: "error",
				Error:  "invalid time_begin format",
			})
			return
		}
	}

	timeEnd, err := time.Parse("2006-01-02T15:04:05Z", timeEndStr)
	if err != nil {
		timeEnd, err = time.Parse("2006-01-02", timeEndStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Status: "error",
				Error:  "invalid time_end format",
			})
			return
		}
	}

	sites := parseIntArray(c.Query("sites"))
	indicators := parseStringArray(c.Query("indicators"))

	data, err := h.service.GetStatistics(code, timeBegin, timeEnd, sites, indicators)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// HealthCheck handles GET /api/health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data: map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		},
	})
}

// DebugArchiveCheck handles GET /api/debug/archive-check
func (h *Handler) DebugArchiveCheck(c *gin.Context) {
	dateStr := c.DefaultQuery("date", "2026-04-01")
	interval := c.DefaultQuery("interval", "day")

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  "invalid date format, use YYYY-MM-DD",
		})
		return
	}

	timeBegin := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	timeEnd := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, time.UTC)

	// Get API key from environment variable
	apiKey := os.Getenv("SENSOR_API_KEY")
	if apiKey == "" {
		apiKey = "not_configured"
	}

	apiKeyMasked := "***"
	if len(apiKey) > 4 {
		apiKeyMasked = "***" + apiKey[len(apiKey)-4:]
	}

	// Call service to get data
	data, err := h.service.GetAggregatedData("knc-air", timeBegin, timeEnd, interval, nil, nil)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Status: "success",
			Data: map[string]interface{}{
				"api_key_exists":   apiKey != "",
				"api_key_length":   len(apiKey),
				"api_key_last4":    apiKeyMasked,
				"time_begin":       timeBegin.Format(time.RFC3339),
				"time_end":         timeEnd.Format(time.RFC3339),
				"interval":         interval,
				"external_api_url": "http://sensor.krasn.ru/hub/api/3.0/sets/knc-air/data/archive",
				"error":            err.Error(),
				"data_length":      0,
				"first_records":    []interface{}{},
			},
		})
		return
	}

	// Try to parse data to get length
	dataLength := 0
	firstRecords := []interface{}{}
	rawDataType := fmt.Sprintf("%T", data)

	// Check if data is json.RawMessage
	if rawMsg, ok := data.(json.RawMessage); ok {
		var records []map[string]interface{}
		if err := json.Unmarshal(rawMsg, &records); err == nil {
			dataLength = len(records)
			if len(records) > 0 {
				firstRecords = append(firstRecords, records[0])
			}
			if len(records) > 1 {
				firstRecords = append(firstRecords, records[1])
			}
		}
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data: map[string]interface{}{
			"api_key_exists":   apiKey != "",
			"api_key_length":   len(apiKey),
			"api_key_last4":    apiKeyMasked,
			"time_begin":       timeBegin.Format(time.RFC3339),
			"time_end":         timeEnd.Format(time.RFC3339),
			"interval":         interval,
			"external_api_url": "http://sensor.krasn.ru/hub/api/3.0/sets/knc-air/data/archive",
			"data_length":      dataLength,
			"data_type":        rawDataType,
			"first_records":    firstRecords,
		},
	})
}

// parseIntArray parses comma-separated integers
func parseIntArray(s string) []int {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if num, err := strconv.Atoi(part); err == nil {
			result = append(result, num)
		}
	}

	return result
}

// parseStringArray parses comma-separated strings
func parseStringArray(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
