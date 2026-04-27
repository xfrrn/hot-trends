package models

import "time"

// TrendItem represents a single hot trend item
type TrendItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Desc      string `json:"desc,omitempty"`
	HotValue  int64  `json:"hot_value,omitempty"`
	URL       string `json:"url"`
	MobileURL string `json:"mobile_url,omitempty"`
	Pic       string `json:"pic,omitempty"`
	Label     string `json:"label,omitempty"`
}

// PlatformTrends represents trends from a single platform
type PlatformTrends struct {
	Platform  string      `json:"platform"`
	Items     []TrendItem `json:"items"`
	FetchedAt time.Time   `json:"fetched_at"`
	Cached    bool        `json:"cached"`
	Error     string      `json:"error,omitempty"`
}

// BatchRequest represents a batch trends request
type BatchRequest struct {
	Platforms      []string `json:"platforms" binding:"required"`
	Limit          int      `json:"limit"`
	Keyword        string   `json:"keyword"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

// BatchResponse represents a batch trends response
type BatchResponse struct {
	Results         []*PlatformTrends `json:"results"`
	TotalPlatforms  int               `json:"total_platforms"`
	Successful      int               `json:"successful"`
	Failed          int               `json:"failed"`
}

// PlatformInfo represents metadata about a platform
type PlatformInfo struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Type                   string `json:"type"`
	RateLimit              string `json:"rate_limit"`
	CacheTTL               string `json:"cache_ttl"`
	Enabled                bool   `json:"enabled"`
	RequiresSpecialHeaders bool   `json:"requires_special_headers,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status            string    `json:"status"`
	Timestamp         time.Time `json:"timestamp"`
	RedisConnected    bool      `json:"redis_connected"`
	ScrapersRegistered int      `json:"scrapers_registered"`
}
