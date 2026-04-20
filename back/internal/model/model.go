package model

import "time"

// APIResponse represents the standard API response structure
type APIResponse struct {
	Status APIStatus   `json:"status"`
	Data   interface{} `json:"data"`
}

// APIStatus represents the status part of API response
type APIStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// DataSet represents a dataset from the external API
type DataSet struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	ShortName   string    `json:"short_name,omitempty"`
	Description string    `json:"description,omitempty"`
	LastTime    time.Time `json:"last_time"`
}

// Site represents a monitoring site
type Site struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Timing  string  `json:"timing,omitempty"`
	GeomX   float64 `json:"geom_x"` // longitude
	GeomY   float64 `json:"geom_y"` // latitude
	GeomAlt *float64 `json:"geom_alt,omitempty"` // altitude
	GeomDir *float64 `json:"geom_dir,omitempty"` // direction
}

// Indicator represents a measurement indicator
type Indicator struct {
	ID    int    `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Units string `json:"units"`
	Tag   *int   `json:"tag,omitempty"`
}

// DataSetDetail represents detailed dataset information
type DataSetDetail struct {
	ID          int                    `json:"id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	ShortName   string                 `json:"short_name,omitempty"`
	Description string                 `json:"description,omitempty"`
	LastTime    time.Time              `json:"last_time"`
	PeriodBegin time.Time              `json:"period_begin"`
	PeriodEnd   *time.Time             `json:"period_end"`
	PeriodLimit map[string]int         `json:"period_limit"`
	Sites       []Site                 `json:"sites"`
	Indicators  []Indicator            `json:"indicators"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// SensorData represents current sensor data
type SensorData struct {
	Site int                    `json:"site"`
	Time time.Time              `json:"time"`
	Data map[string]interface{} `json:"-"` // Dynamic fields for indicators
}

// RawData represents raw sensor data
type RawData struct {
	Time time.Time              `json:"time"`
	Site int                    `json:"site"`
	Data map[string]interface{} `json:"-"`
}

// AggregatedData represents aggregated sensor data
type AggregatedData struct {
	Time time.Time              `json:"time"`
	Site int                    `json:"site"`
	Data map[string]interface{} `json:"-"`
}

// StatData represents statistical data
type StatData struct {
	Avg float64 `json:"avg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Std float64 `json:"std"`
	Cnt int     `json:"cnt"`
}

// AggregatedDataExt represents aggregated data with statistics
type AggregatedDataExt struct {
	Time time.Time            `json:"time"`
	Site int                  `json:"site"`
	Data map[string]*StatData `json:"-"`
}

// TimeInterval represents time aggregation interval
type TimeInterval string

const (
	IntervalHour  TimeInterval = "hour"
	IntervalDay   TimeInterval = "day"
	IntervalMonth TimeInterval = "month"
	IntervalYear  TimeInterval = "year"
)

// QueryParams represents common query parameters
type QueryParams struct {
	TimeBegin    *time.Time   `form:"time_begin"`
	TimeEnd      *time.Time   `form:"time_end"`
	TimeInterval TimeInterval `form:"time_interval"`
	Sites        []int        `form:"sites"`
	Indicators   []string     `form:"indicators"`
}
