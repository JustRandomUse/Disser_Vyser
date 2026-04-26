package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
	once   sync.Once
)

// Client for external API
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func newClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 180 * time.Second},
		baseURL:    "http://sensor.krasn.ru/hub/api/3.0",
		apiKey:     apiKey,
	}
}

func (c *Client) doRequest(method, path string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("uid", c.apiKey)
	params.Set("opt-date", "odbc")

	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return body, nil
}

func (c *Client) getArchiveData(setCode string, timeBegin, timeEnd time.Time, interval string) ([]byte, error) {
	path := fmt.Sprintf("/sets/%s/data/archive", setCode)
	params := url.Values{}
	params.Set("time_begin", timeBegin.Format("2006-01-02 15:04:05"))
	params.Set("time_end", timeEnd.Format("2006-01-02 15:04:05"))
	params.Set("time_interval", interval)
	return c.doRequest("GET", path, params)
}

func parseAPIResponse(data []byte) (json.RawMessage, error) {
	var response struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if response.Status.Code < 0 {
		return nil, fmt.Errorf("API error: %s", response.Status.Message)
	}

	return response.Data, nil
}

type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func initRouter() *gin.Engine {
	once.Do(func() {
		apiKey := os.Getenv("SENSOR_API_KEY")
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "https://disser-front.vercel.app,http://localhost:5173"
		}

		client := newClient(apiKey)

		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.Use(gin.Recovery())

		config := cors.DefaultConfig()
		config.AllowOrigins = strings.Split(allowedOrigins, ",")
		config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
		config.AllowCredentials = true
		r.Use(cors.New(config))

		api := r.Group("/api")
		{
			api.GET("/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, Response{
					Status: "success",
					Data: map[string]string{
						"status": "healthy",
						"time":   time.Now().Format(time.RFC3339),
					},
				})
			})

			api.GET("/datasets/knc-air/aggregated", func(c *gin.Context) {
				timeBeginStr := c.Query("time_begin")
				timeEndStr := c.Query("time_end")
				interval := c.DefaultQuery("interval", "day")

				timeBegin, err := time.Parse("2006-01-02T15:04:05Z", timeBeginStr)
				if err != nil {
					timeBegin, _ = time.Parse("2006-01-02", timeBeginStr)
				}

				timeEnd, err := time.Parse("2006-01-02T15:04:05Z", timeEndStr)
				if err != nil {
					timeEnd, _ = time.Parse("2006-01-02", timeEndStr)
				}

				data, err := client.getArchiveData("knc-air", timeBegin, timeEnd, interval)
				if err != nil {
					c.JSON(http.StatusInternalServerError, Response{
						Status: "error",
						Error:  err.Error(),
					})
					return
				}

				result, err := parseAPIResponse(data)
				if err != nil {
					c.JSON(http.StatusInternalServerError, Response{
						Status: "error",
						Error:  err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, Response{
					Status: "success",
					Data:   result,
				})
			})

			api.GET("/debug/archive-check", func(c *gin.Context) {
				dateStr := c.DefaultQuery("date", "2026-04-01")
				interval := c.DefaultQuery("interval", "day")

				date, _ := time.Parse("2006-01-02", dateStr)
				timeBegin := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
				timeEnd := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, time.UTC)

				data, err := client.getArchiveData("knc-air", timeBegin, timeEnd, interval)
				if err != nil {
					c.JSON(http.StatusOK, Response{
						Status: "success",
						Data: map[string]interface{}{
							"error": err.Error(),
						},
					})
					return
				}

				result, _ := parseAPIResponse(data)
				var records []map[string]interface{}
				json.Unmarshal(result, &records)

				c.JSON(http.StatusOK, Response{
					Status: "success",
					Data: map[string]interface{}{
						"data_length":   len(records),
						"time_begin":    timeBegin.Format(time.RFC3339),
						"time_end":      timeEnd.Format(time.RFC3339),
						"interval":      interval,
						"first_records": records[:min(2, len(records))],
					},
				})
			})
		}

		router = r
	})

	return router
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	initRouter().ServeHTTP(w, r)
}
