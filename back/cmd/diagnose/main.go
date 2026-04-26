package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	DirectAPIBase = "http://sensor.krasn.ru/hub/api/3.0"
	BackendBase   = "http://localhost:8080/api"
)

type ComparisonResult struct {
	DirectURL      string
	BackendURL     string
	DirectStatus   int
	BackendStatus  int
	DirectSize     int
	BackendSize    int
	DirectCount    int
	BackendCount   int
	DirectError    string
	BackendError   string
	DirectData     interface{}
	BackendData    interface{}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <api_key>")
		fmt.Println("Example: go run main.go your-api-key-here")
		os.Exit(1)
	}

	apiKey := os.Args[1]

	// Test case: 2026-04-24 data
	testDate := "2026-04-24"
	testTimeBegin := "2026-04-24 00:00:00"
	testTimeEnd := "2026-04-24 23:59:59"

	fmt.Println("=== Air Quality Monitor API Diagnostic Tool ===")
	fmt.Printf("Testing date: %s\n\n", testDate)

	// Test 1: Archive data (hourly)
	fmt.Println("--- Test 1: Archive Data (hourly) ---")
	result1 := compareArchiveData(apiKey, testTimeBegin, testTimeEnd, "hour")
	printComparison(result1)

	// Test 2: Archive data (daily)
	fmt.Println("\n--- Test 2: Archive Data (daily) ---")
	result2 := compareArchiveData(apiKey, testTimeBegin, testTimeEnd, "day")
	printComparison(result2)

	// Test 3: Last data
	fmt.Println("\n--- Test 3: Last Data ---")
	result3 := compareLastData(apiKey)
	printComparison(result3)
}

func compareArchiveData(apiKey, timeBegin, timeEnd, interval string) ComparisonResult {
	result := ComparisonResult{}

	// Direct API call
	params := url.Values{}
	params.Set("uid", apiKey)
	params.Set("opt-date", "odbc")
	params.Set("time_begin", timeBegin)
	params.Set("time_end", timeEnd)
	params.Set("time_interval", interval)

	directURL := fmt.Sprintf("%s/sets/knc-air/data/archive?%s", DirectAPIBase, params.Encode())
	result.DirectURL = maskAPIKey(directURL, apiKey)

	directResp, directBody, directErr := makeRequest(directURL)
	if directErr != nil {
		result.DirectError = directErr.Error()
	} else {
		result.DirectStatus = directResp.StatusCode
		result.DirectSize = len(directBody)
		result.DirectCount = countDataRecords(directBody)
		result.DirectData = parseResponse(directBody)
	}

	// Backend API call
	backendParams := url.Values{}
	backendParams.Set("time_begin", timeBegin)
	backendParams.Set("time_end", timeEnd)
	backendParams.Set("interval", interval)

	backendURL := fmt.Sprintf("%s/datasets/knc-air/aggregated?%s", BackendBase, backendParams.Encode())
	result.BackendURL = backendURL

	backendResp, backendBody, backendErr := makeRequest(backendURL)
	if backendErr != nil {
		result.BackendError = backendErr.Error()
	} else {
		result.BackendStatus = backendResp.StatusCode
		result.BackendSize = len(backendBody)
		result.BackendCount = countDataRecords(backendBody)
		result.BackendData = parseResponse(backendBody)
	}

	return result
}

func compareLastData(apiKey string) ComparisonResult {
	result := ComparisonResult{}

	// Direct API call
	params := url.Values{}
	params.Set("uid", apiKey)
	params.Set("opt-date", "odbc")

	directURL := fmt.Sprintf("%s/sets/knc-air/data/last?%s", DirectAPIBase, params.Encode())
	result.DirectURL = maskAPIKey(directURL, apiKey)

	directResp, directBody, directErr := makeRequest(directURL)
	if directErr != nil {
		result.DirectError = directErr.Error()
	} else {
		result.DirectStatus = directResp.StatusCode
		result.DirectSize = len(directBody)
		result.DirectCount = countDataRecords(directBody)
		result.DirectData = parseResponse(directBody)
	}

	// Backend API call
	backendURL := fmt.Sprintf("%s/datasets/knc-air/last", BackendBase)
	result.BackendURL = backendURL

	backendResp, backendBody, backendErr := makeRequest(backendURL)
	if backendErr != nil {
		result.BackendError = backendErr.Error()
	} else {
		result.BackendStatus = backendResp.StatusCode
		result.BackendSize = len(backendBody)
		result.BackendCount = countDataRecords(backendBody)
		result.BackendData = parseResponse(backendBody)
	}

	return result
}

func makeRequest(url string) (*http.Response, []byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}

	return resp, body, nil
}

func parseResponse(body []byte) interface{} {
	var response struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil
	}

	return response.Data
}

func countDataRecords(body []byte) int {
	var response struct {
		Status struct {
			Code int `json:"code"`
		} `json:"status"`
		Data []interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return -1
	}

	return len(response.Data)
}

func maskAPIKey(url, apiKey string) string {
	if len(apiKey) > 4 {
		masked := "***" + apiKey[len(apiKey)-4:]
		return fmt.Sprintf("%s (uid=%s)", url[:len(url)-len(apiKey)-4], masked)
	}
	return url
}

func printComparison(result ComparisonResult) {
	fmt.Printf("Direct API URL: %s\n", result.DirectURL)
	fmt.Printf("Backend URL:    %s\n\n", result.BackendURL)

	if result.DirectError != "" {
		fmt.Printf("❌ Direct API Error: %s\n", result.DirectError)
	} else {
		fmt.Printf("✅ Direct API Status: %d\n", result.DirectStatus)
		fmt.Printf("   Response Size: %d bytes\n", result.DirectSize)
		fmt.Printf("   Data Records: %d\n", result.DirectCount)
	}

	fmt.Println()

	if result.BackendError != "" {
		fmt.Printf("❌ Backend Error: %s\n", result.BackendError)
	} else {
		fmt.Printf("✅ Backend Status: %d\n", result.BackendStatus)
		fmt.Printf("   Response Size: %d bytes\n", result.BackendSize)
		fmt.Printf("   Data Records: %d\n", result.BackendCount)
	}

	fmt.Println()

	// Comparison
	if result.DirectError == "" && result.BackendError == "" {
		if result.DirectCount == result.BackendCount {
			fmt.Printf("✅ Record counts match: %d\n", result.DirectCount)
		} else {
			fmt.Printf("⚠️  Record count mismatch: Direct=%d, Backend=%d\n", result.DirectCount, result.BackendCount)
		}

		if result.DirectSize == result.BackendSize {
			fmt.Printf("✅ Response sizes match: %d bytes\n", result.DirectSize)
		} else {
			fmt.Printf("⚠️  Response size difference: Direct=%d bytes, Backend=%d bytes\n", result.DirectSize, result.BackendSize)
		}
	}

	// Show sample data if counts differ
	if result.DirectCount != result.BackendCount && result.DirectCount > 0 {
		fmt.Println("\n📊 Sample Direct API Data (first record):")
		if data, ok := result.DirectData.(json.RawMessage); ok {
			var records []map[string]interface{}
			if err := json.Unmarshal(data, &records); err == nil && len(records) > 0 {
				prettyPrint(records[0])
			}
		}
	}

	if result.BackendCount > 0 {
		fmt.Println("\n📊 Sample Backend Data (first record):")
		if data, ok := result.BackendData.(json.RawMessage); ok {
			var records []map[string]interface{}
			if err := json.Unmarshal(data, &records); err == nil && len(records) > 0 {
				prettyPrint(records[0])
			}
		}
	}
}

func prettyPrint(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "   ", "  ")
	if err != nil {
		fmt.Printf("   %v\n", data)
		return
	}
	fmt.Printf("   %s\n", string(jsonBytes))
}
