package aggregator

import (
	"encoding/json"
	"math"
	"time"
)

// AggregateData aggregates raw data by time interval
func AggregateData(rawData []map[string]interface{}, interval string) ([]map[string]interface{}, error) {
	if len(rawData) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Group data by time intervals and sites
	groups := make(map[string]map[int][]map[string]interface{})

	for _, record := range rawData {
		timeStr, ok := record["time"].(string)
		if !ok {
			continue
		}

		t, err := time.Parse("2006-01-02 15:04:05", timeStr)
		if err != nil {
			continue
		}

		// Round time to interval
		intervalKey := roundTimeToInterval(t, interval)

		siteID, ok := record["site"].(float64)
		if !ok {
			continue
		}
		site := int(siteID)

		if groups[intervalKey] == nil {
			groups[intervalKey] = make(map[int][]map[string]interface{})
		}

		groups[intervalKey][site] = append(groups[intervalKey][site], record)
	}

	// Aggregate each group
	result := make([]map[string]interface{}, 0)

	for intervalKey, siteGroups := range groups {
		for site, records := range siteGroups {
			aggregated := aggregateRecords(records, intervalKey, site)
			result = append(result, aggregated)
		}
	}

	return result, nil
}

// roundTimeToInterval rounds time to the specified interval
func roundTimeToInterval(t time.Time, interval string) string {
	switch interval {
	case "hour":
		return t.Truncate(time.Hour).Format("2006-01-02 15:04:05")
	case "day":
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Format("2006-01-02 15:04:05")
	case "month":
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).Format("2006-01-02 15:04:05")
	case "year":
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()).Format("2006-01-02 15:04:05")
	default:
		return t.Format("2006-01-02 15:04:05")
	}
}

// aggregateRecords aggregates multiple records into one
func aggregateRecords(records []map[string]interface{}, timeKey string, site int) map[string]interface{} {
	result := map[string]interface{}{
		"time": timeKey,
		"site": site,
	}

	// Collect all indicator values
	indicators := make(map[string][]float64)

	for _, record := range records {
		for key, value := range record {
			if key == "time" || key == "site" {
				continue
			}

			// Try to convert to float64
			var floatVal float64
			switch v := value.(type) {
			case float64:
				floatVal = v
			case int:
				floatVal = float64(v)
			case json.Number:
				f, err := v.Float64()
				if err == nil {
					floatVal = f
				}
			default:
				continue
			}

			indicators[key] = append(indicators[key], floatVal)
		}
	}

	// Calculate averages
	for indicator, values := range indicators {
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			result[indicator] = math.Round(sum/float64(len(values))*10) / 10
		}
	}

	return result
}

// AggregateDataWithStats aggregates data and calculates statistics
func AggregateDataWithStats(rawData []map[string]interface{}, interval string) ([]map[string]interface{}, error) {
	if len(rawData) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Group data by time intervals and sites
	groups := make(map[string]map[int][]map[string]interface{})

	for _, record := range rawData {
		timeStr, ok := record["time"].(string)
		if !ok {
			continue
		}

		t, err := time.Parse("2006-01-02 15:04:05", timeStr)
		if err != nil {
			continue
		}

		intervalKey := roundTimeToInterval(t, interval)

		siteID, ok := record["site"].(float64)
		if !ok {
			continue
		}
		site := int(siteID)

		if groups[intervalKey] == nil {
			groups[intervalKey] = make(map[int][]map[string]interface{})
		}

		groups[intervalKey][site] = append(groups[intervalKey][site], record)
	}

	// Aggregate each group with statistics
	result := make([]map[string]interface{}, 0)

	for intervalKey, siteGroups := range groups {
		for site, records := range siteGroups {
			aggregated := aggregateRecordsWithStats(records, intervalKey, site)
			result = append(result, aggregated)
		}
	}

	return result, nil
}

// aggregateRecordsWithStats aggregates records and calculates statistics
func aggregateRecordsWithStats(records []map[string]interface{}, timeKey string, site int) map[string]interface{} {
	result := map[string]interface{}{
		"time": timeKey,
		"site": site,
	}

	// Collect all indicator values
	indicators := make(map[string][]float64)

	for _, record := range records {
		for key, value := range record {
			if key == "time" || key == "site" {
				continue
			}

			var floatVal float64
			switch v := value.(type) {
			case float64:
				floatVal = v
			case int:
				floatVal = float64(v)
			case json.Number:
				f, err := v.Float64()
				if err == nil {
					floatVal = f
				}
			default:
				continue
			}

			indicators[key] = append(indicators[key], floatVal)
		}
	}

	// Calculate statistics for each indicator
	for indicator, values := range indicators {
		if len(values) > 0 {
			stats := calculateStats(values)
			result[indicator] = stats
		}
	}

	return result
}

// calculateStats calculates statistical measures
func calculateStats(values []float64) map[string]interface{} {
	if len(values) == 0 {
		return nil
	}

	sum := 0.0
	min := values[0]
	max := values[0]

	for _, v := range values {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	avg := sum / float64(len(values))

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		diff := v - avg
		variance += diff * diff
	}
	variance /= float64(len(values))
	std := math.Sqrt(variance)

	return map[string]interface{}{
		"avg": math.Round(avg*10) / 10,
		"min": math.Round(min*10) / 10,
		"max": math.Round(max*10) / 10,
		"std": math.Round(std*10) / 10,
		"cnt": len(values),
	}
}

// CalculateOverallStats calculates overall statistics for a dataset
func CalculateOverallStats(data []map[string]interface{}) map[string]interface{} {
	// Group by site
	siteData := make(map[int][]map[string]interface{})

	for _, record := range data {
		siteID, ok := record["site"].(float64)
		if !ok {
			continue
		}
		site := int(siteID)

		siteData[site] = append(siteData[site], record)
	}

	// Calculate stats for each site
	result := make([]map[string]interface{}, 0)

	for site, records := range siteData {
		siteStats := map[string]interface{}{
			"site": site,
		}

		// Collect all indicator values
		indicators := make(map[string][]float64)

		for _, record := range records {
			for key, value := range record {
				if key == "time" || key == "site" {
					continue
				}

				var floatVal float64
				switch v := value.(type) {
				case float64:
					floatVal = v
				case int:
					floatVal = float64(v)
				case json.Number:
					f, err := v.Float64()
					if err == nil {
						floatVal = f
					}
				default:
					continue
				}

				indicators[key] = append(indicators[key], floatVal)
			}
		}

		// Calculate stats for each indicator
		for indicator, values := range indicators {
			if len(values) > 0 {
				stats := calculateStats(values)
				siteStats[indicator] = stats
			}
		}

		result = append(result, siteStats)
	}

	return map[string]interface{}{
		"data": result,
	}
}
