package service

import (
	"air-quality-monitor/back/internal/aggregator"
	"air-quality-monitor/back/internal/cache"
	"air-quality-monitor/back/internal/client"
	"encoding/json"
	"fmt"
	"time"
)

// Service handles business logic
type Service struct {
	client *client.Client
	cache  *cache.Cache
}

// NewService creates a new service instance
func NewService(apiKey string) *Service {
	return &Service{
		client: client.NewClient(apiKey),
		cache:  cache.NewCache(),
	}
}

// GetDataSets retrieves list of datasets
func (s *Service) GetDataSets() (interface{}, error) {
	cacheKey := "datasets"

	// Check cache
	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Fetch from API
	data, err := s.client.GetDataSets()
	if err != nil {
		return nil, err
	}

	result, err := client.ParseAPIResponse(data)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	s.cache.Set(cacheKey, result, 5*time.Minute)

	return result, nil
}

// GetDataSetDetail retrieves dataset details
func (s *Service) GetDataSetDetail(setCode string) (interface{}, error) {
	cacheKey := fmt.Sprintf("dataset:%s", setCode)

	// Check cache
	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Fetch from API
	data, err := s.client.GetDataSetDetail(setCode)
	if err != nil {
		return nil, err
	}

	result, err := client.ParseAPIResponse(data)
	if err != nil {
		return nil, err
	}

	// Cache for 10 minutes
	s.cache.Set(cacheKey, result, 10*time.Minute)

	return result, nil
}

// GetLastData retrieves current sensor data
func (s *Service) GetLastData(setCode string, sites []int, indicators []string) (interface{}, error) {
	cacheKey := fmt.Sprintf("last:%s:%v:%v", setCode, sites, indicators)

	// Check cache (short TTL for live data)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Fetch from API
	data, err := s.client.GetLastData(setCode, sites, indicators)
	if err != nil {
		return nil, err
	}

	result, err := client.ParseAPIResponse(data)
	if err != nil {
		return nil, err
	}

	// Cache for 1 minute (live data)
	s.cache.Set(cacheKey, result, 1*time.Minute)

	return result, nil
}

// GetAggregatedData retrieves and aggregates data
func (s *Service) GetAggregatedData(setCode string, timeBegin, timeEnd time.Time, interval string, sites []int, indicators []string) (interface{}, error) {
	cacheKey := fmt.Sprintf("agg:%s:%s:%s:%s:%v:%v", setCode, timeBegin.Format("2006-01-02 15:04:05"), timeEnd.Format("2006-01-02 15:04:05"), interval, sites, indicators)

	// Check cache
	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Determine if we should use external API aggregation or do it ourselves
	duration := timeEnd.Sub(timeBegin)

	// For large time ranges, use external API aggregation
	if duration > 7*24*time.Hour || interval == "day" || interval == "month" {
		data, err := s.client.GetArchiveData(setCode, timeBegin, timeEnd, interval, sites, indicators)
		if err != nil {
			return nil, err
		}

		result, err := client.ParseAPIResponse(data)
		if err != nil {
			return nil, err
		}

		// Cache for longer (historical data doesn't change)
		cacheDuration := 30 * time.Minute
		if duration > 30*24*time.Hour {
			cacheDuration = 2 * time.Hour
		}

		s.cache.Set(cacheKey, result, cacheDuration)
		return result, nil
	}

	// For smaller ranges, fetch raw data and aggregate on backend
	data, err := s.client.GetRawData(setCode, timeBegin, timeEnd, sites, indicators)
	if err != nil {
		return nil, err
	}

	rawResult, err := client.ParseAPIResponse(data)
	if err != nil {
		return nil, err
	}

	// Parse raw data
	var rawData []map[string]interface{}
	if err := json.Unmarshal(rawResult.(json.RawMessage), &rawData); err != nil {
		return nil, err
	}

	// Aggregate data
	aggregated, err := aggregator.AggregateData(rawData, interval)
	if err != nil {
		return nil, err
	}

	// Cache aggregated result
	s.cache.Set(cacheKey, aggregated, 15*time.Minute)

	return aggregated, nil
}

// GetAggregatedDataWithStats retrieves aggregated data with statistics
func (s *Service) GetAggregatedDataWithStats(setCode string, timeBegin, timeEnd time.Time, interval string, sites []int, indicators []string) (interface{}, error) {
	cacheKey := fmt.Sprintf("agg-stats:%s:%s:%s:%s:%v:%v", setCode, timeBegin.Format("2006-01-02 15:04:05"), timeEnd.Format("2006-01-02 15:04:05"), interval, sites, indicators)

	// Check cache
	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Use external API for statistics
	data, err := s.client.GetArchiveDataExt(setCode, timeBegin, timeEnd, interval, sites, indicators)
	if err != nil {
		return nil, err
	}

	result, err := client.ParseAPIResponse(data)
	if err != nil {
		return nil, err
	}

	// Cache for longer (historical data with stats)
	duration := timeEnd.Sub(timeBegin)
	cacheDuration := 30 * time.Minute
	if duration > 30*24*time.Hour {
		cacheDuration = 2 * time.Hour
	}

	s.cache.Set(cacheKey, result, cacheDuration)

	return result, nil
}

// GetTimeSeriesData retrieves time series data for charts
func (s *Service) GetTimeSeriesData(setCode string, timeBegin, timeEnd time.Time, interval string, sites []int, indicators []string) (interface{}, error) {
	// Optimize interval based on time range
	duration := timeEnd.Sub(timeBegin)

	if interval == "" {
		if duration <= 24*time.Hour {
			interval = "hour"
		} else if duration <= 30*24*time.Hour {
			interval = "day"
		} else {
			interval = "month"
		}
	}

	return s.GetAggregatedData(setCode, timeBegin, timeEnd, interval, sites, indicators)
}

// GetStatistics calculates overall statistics for selected sites
func (s *Service) GetStatistics(setCode string, timeBegin, timeEnd time.Time, sites []int, indicators []string) (interface{}, error) {
	cacheKey := fmt.Sprintf("stats:%s:%s:%s:%v:%v", setCode, timeBegin.Format("2006-01-02 15:04:05"), timeEnd.Format("2006-01-02 15:04:05"), sites, indicators)

	// Check cache
	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Determine optimal interval
	duration := timeEnd.Sub(timeBegin)
	interval := "hour"
	if duration > 7*24*time.Hour {
		interval = "day"
	}

	// Get aggregated data with stats
	data, err := s.client.GetArchiveDataExt(setCode, timeBegin, timeEnd, interval, sites, indicators)
	if err != nil {
		return nil, err
	}

	result, err := client.ParseAPIResponse(data)
	if err != nil {
		return nil, err
	}

	// Parse and calculate overall statistics
	var aggData []map[string]interface{}
	if err := json.Unmarshal(result.(json.RawMessage), &aggData); err != nil {
		return nil, err
	}

	stats := aggregator.CalculateOverallStats(aggData)

	// Cache statistics
	s.cache.Set(cacheKey, stats, 30*time.Minute)

	return stats, nil
}

// ClearCache clears all cached data
func (s *Service) ClearCache() {
	s.cache.Clear()
}
