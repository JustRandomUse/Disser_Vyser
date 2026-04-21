package handler

import (
	"air-quality-monitor/back/internal/service"
	"net/http"
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
	} else if now.Sub(timeBegin) < 24*time.Hour {
		// Use raw data for recent hours (< 24h ago)
		data, err = h.service.GetAggregatedData(code, timeBegin, timeEnd, "hour", sites, indicators)
	} else {
		// Use archive data for older data (> 24h ago)
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
