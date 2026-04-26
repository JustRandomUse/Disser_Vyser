package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURL        = "http://sensor.krasn.ru/hub/api/3.0"
	DefaultTimeout = 180 * time.Second
	MaxRetries     = 1
	RetryDelay     = 2 * time.Second
)

// Client represents the external API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewClient creates a new API client
func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: BaseURL,
		apiKey:  apiKey,
	}
}

// doRequest performs HTTP request with retry logic
func (c *Client) doRequest(method, path string, params url.Values) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(RetryDelay)
		}

		// Add API key to params
		if params == nil {
			params = url.Values{}
		}
		params.Set("uid", c.apiKey)
		params.Set("opt-date", "odbc")

		reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

		// Log the actual request URL (mask API key for security)
		maskedURL := reqURL
		if len(c.apiKey) > 4 {
			maskedURL = fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())
			maskedURL = fmt.Sprintf("%s (uid=***%s)", maskedURL[:len(maskedURL)-len(c.apiKey)-4], c.apiKey[len(c.apiKey)-4:])
		}
		fmt.Printf("🔗 External API Request: %s\n", maskedURL)

		req, err := http.NewRequest(method, reqURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("❌ External API returned status %d: %s\n", resp.StatusCode, string(body))
			lastErr = fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
			continue
		}

		fmt.Printf("✅ External API response: %d bytes\n", len(body))
		return body, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", MaxRetries, lastErr)
}

// GetDataSets retrieves list of available datasets
func (c *Client) GetDataSets() ([]byte, error) {
	return c.doRequest("GET", "/sets", nil)
}

// GetDataSetDetail retrieves detailed information about a dataset
func (c *Client) GetDataSetDetail(setCode string) ([]byte, error) {
	path := fmt.Sprintf("/sets/%s", setCode)
	return c.doRequest("GET", path, nil)
}

// GetLastData retrieves current/latest data from sensors
func (c *Client) GetLastData(setCode string, sites []int, indicators []string) ([]byte, error) {
	path := fmt.Sprintf("/sets/%s/data/last", setCode)
	params := url.Values{}

	if len(sites) > 0 {
		siteStr := ""
		for i, site := range sites {
			if i > 0 {
				siteStr += ","
			}
			siteStr += fmt.Sprintf("%d", site)
		}
		params.Set("sites", siteStr)
	}

	if len(indicators) > 0 {
		indStr := ""
		for i, ind := range indicators {
			if i > 0 {
				indStr += ","
			}
			indStr += ind
		}
		params.Set("indicators", indStr)
	}

	return c.doRequest("GET", path, params)
}

// GetLastDataExt retrieves current data in extended format
func (c *Client) GetLastDataExt(setCode string, sites []int, indicators []string) ([]byte, error) {
	path := fmt.Sprintf("/sets/%s/data/last-ext", setCode)
	params := url.Values{}

	if len(sites) > 0 {
		siteStr := ""
		for i, site := range sites {
			if i > 0 {
				siteStr += ","
			}
			siteStr += fmt.Sprintf("%d", site)
		}
		params.Set("sites", siteStr)
	}

	if len(indicators) > 0 {
		indStr := ""
		for i, ind := range indicators {
			if i > 0 {
				indStr += ","
			}
			indStr += ind
		}
		params.Set("indicators", indStr)
	}

	return c.doRequest("GET", path, params)
}

// GetRawData retrieves raw sensor data
func (c *Client) GetRawData(setCode string, timeBegin, timeEnd time.Time, sites []int, indicators []string) ([]byte, error) {
	path := fmt.Sprintf("/sets/%s/data/raw", setCode)
	params := url.Values{}

	params.Set("time_begin", timeBegin.Format("2006-01-02 15:04:05"))
	params.Set("time_end", timeEnd.Format("2006-01-02 15:04:05"))

	if len(sites) > 0 {
		siteStr := ""
		for i, site := range sites {
			if i > 0 {
				siteStr += ","
			}
			siteStr += fmt.Sprintf("%d", site)
		}
		params.Set("sites", siteStr)
	}

	if len(indicators) > 0 {
		indStr := ""
		for i, ind := range indicators {
			if i > 0 {
				indStr += ","
			}
			indStr += ind
		}
		params.Set("indicators", indStr)
	}

	return c.doRequest("GET", path, params)
}

// GetArchiveData retrieves aggregated archive data
func (c *Client) GetArchiveData(setCode string, timeBegin, timeEnd time.Time, interval string, sites []int, indicators []string) ([]byte, error) {
	path := fmt.Sprintf("/sets/%s/data/archive", setCode)
	params := url.Values{}

	params.Set("time_begin", timeBegin.Format("2006-01-02 15:04:05"))
	params.Set("time_end", timeEnd.Format("2006-01-02 15:04:05"))
	params.Set("time_interval", interval)

	if len(sites) > 0 {
		siteStr := ""
		for i, site := range sites {
			if i > 0 {
				siteStr += ","
			}
			siteStr += fmt.Sprintf("%d", site)
		}
		params.Set("sites", siteStr)
	}

	if len(indicators) > 0 {
		indStr := ""
		for i, ind := range indicators {
			if i > 0 {
				indStr += ","
			}
			indStr += ind
		}
		params.Set("indicators", indStr)
	}

	return c.doRequest("GET", path, params)
}

// GetArchiveDataExt retrieves aggregated data with statistics
func (c *Client) GetArchiveDataExt(setCode string, timeBegin, timeEnd time.Time, interval string, sites []int, indicators []string) ([]byte, error) {
	path := fmt.Sprintf("/sets/%s/data/archive-ext", setCode)
	params := url.Values{}

	params.Set("time_begin", timeBegin.Format("2006-01-02 15:04:05"))
	params.Set("time_end", timeEnd.Format("2006-01-02 15:04:05"))
	params.Set("time_interval", interval)

	if len(sites) > 0 {
		siteStr := ""
		for i, site := range sites {
			if i > 0 {
				siteStr += ","
			}
			siteStr += fmt.Sprintf("%d", site)
		}
		params.Set("sites", siteStr)
	}

	if len(indicators) > 0 {
		indStr := ""
		for i, ind := range indicators {
			if i > 0 {
				indStr += ","
			}
			indStr += ind
		}
		params.Set("indicators", indStr)
	}

	return c.doRequest("GET", path, params)
}

// ParseAPIResponse parses the standard API response
func ParseAPIResponse(data []byte) (interface{}, error) {
	var response struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Status.Code < 0 {
		return nil, fmt.Errorf("API error: %s (code: %d)", response.Status.Message, response.Status.Code)
	}

	return response.Data, nil
}
