package scraper

import (
	"context"
	"hot-trends-service/internal/models"
	"time"
)

// ScraperType represents the type of scraper
type ScraperType string

const (
	JSONAPIScraper   ScraperType = "json_api"
	HTMLScraper      ScraperType = "html_scraping"
)

// FetchOptions contains options for fetching trends
type FetchOptions struct {
	Limit   int
	Keyword string
	Timeout time.Duration
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	CacheTTL          time.Duration
}

// Scraper is the interface that all platform scrapers must implement
type Scraper interface {
	// Name returns the platform identifier
	Name() string

	// DisplayName returns the human-readable platform name
	DisplayName() string

	// Fetch retrieves trends from the platform
	Fetch(ctx context.Context, opts FetchOptions) (*models.PlatformTrends, error)

	// Type returns the scraper type
	Type() ScraperType

	// RateLimit returns the rate limit configuration
	RateLimit() RateLimitConfig
}

// BaseScraper provides common functionality for scrapers
type BaseScraper struct {
	name        string
	displayName string
	scraperType ScraperType
	rateLimit   RateLimitConfig
}

// NewBaseScraper creates a new base scraper
func NewBaseScraper(name, displayName string, scraperType ScraperType, rateLimit RateLimitConfig) *BaseScraper {
	return &BaseScraper{
		name:        name,
		displayName: displayName,
		scraperType: scraperType,
		rateLimit:   rateLimit,
	}
}

// Name returns the platform identifier
func (b *BaseScraper) Name() string {
	return b.name
}

// DisplayName returns the human-readable platform name
func (b *BaseScraper) DisplayName() string {
	return b.displayName
}

// Type returns the scraper type
func (b *BaseScraper) Type() ScraperType {
	return b.scraperType
}

// RateLimit returns the rate limit configuration
func (b *BaseScraper) RateLimit() RateLimitConfig {
	return b.rateLimit
}
