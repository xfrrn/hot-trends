package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"hot-trends-service/internal/models"
	"log"
	"sync"
	"time"
)

// Cache interface for caching trends
type Cache interface {
	Get(platform, keyword string) (*models.PlatformTrends, bool)
	Set(platform, keyword string, trends *models.PlatformTrends, ttl time.Duration)
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	Wait(ctx context.Context, platform string) error
}

// Executor executes scraper operations with caching and rate limiting
type Executor struct {
	registry    *Registry
	cache       Cache
	rateLimiter RateLimiter
}

// NewExecutor creates a new executor
func NewExecutor(registry *Registry, cache Cache, rateLimiter RateLimiter) *Executor {
	return &Executor{
		registry:    registry,
		cache:       cache,
		rateLimiter: rateLimiter,
	}
}

// FetchSingle fetches trends from a single platform
func (e *Executor) FetchSingle(ctx context.Context, platform string, opts FetchOptions) *models.PlatformTrends {
	// Check cache first
	if e.cache != nil {
		if cached, found := e.cache.Get(platform, opts.Keyword); found {
			cached.Cached = true
			return cached
		}
	}

	// Get scraper
	scraper, ok := e.registry.Get(platform)
	if !ok {
		return &models.PlatformTrends{
			Platform:  platform,
			Items:     []models.TrendItem{},
			FetchedAt: time.Now(),
			Error:     "platform not found",
		}
	}

	// Rate limit check
	if e.rateLimiter != nil {
		if err := e.rateLimiter.Wait(ctx, platform); err != nil {
			return &models.PlatformTrends{
				Platform:  platform,
				Items:     []models.TrendItem{},
				FetchedAt: time.Now(),
				Error:     fmt.Sprintf("rate limit: %v", err),
			}
		}
	}

	// Set timeout if not specified
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}

	// Fetch with timeout
	fetchCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	trends, err := scraper.Fetch(fetchCtx, opts)
	if err != nil {
		log.Printf("fetch failed for %s: %v", platform, err)
		return &models.PlatformTrends{
			Platform:  platform,
			Items:     []models.TrendItem{},
			FetchedAt: time.Now(),
			Error:     err.Error(),
		}
	}

	// Cache result
	if e.cache != nil && trends.Error == "" {
		e.cache.Set(platform, opts.Keyword, trends, scraper.RateLimit().CacheTTL)
	}

	return trends
}

// FetchMultiple fetches trends from multiple platforms concurrently
func (e *Executor) FetchMultiple(ctx context.Context, platforms []string, opts FetchOptions) []*models.PlatformTrends {
	results := make([]*models.PlatformTrends, len(platforms))
	var wg sync.WaitGroup

	for i, platform := range platforms {
		wg.Add(1)
		go func(idx int, plat string) {
			defer wg.Done()
			results[idx] = e.FetchSingle(ctx, plat, opts)
		}(i, platform)
	}

	wg.Wait()
	return results
}

// MemoryCache is a simple in-memory cache implementation
type MemoryCache struct {
	data map[string]cacheEntry
	mu   sync.RWMutex
}

type cacheEntry struct {
	trends    *models.PlatformTrends
	expiresAt time.Time
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

func (m *MemoryCache) buildKey(platform, keyword string) string {
	if keyword == "" {
		return platform
	}
	return fmt.Sprintf("%s:%s", platform, keyword)
}

func (m *MemoryCache) Get(platform, keyword string) (*models.PlatformTrends, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.buildKey(platform, keyword)
	entry, ok := m.data[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	// Deep copy to avoid mutation
	trendsJSON, _ := json.Marshal(entry.trends)
	var trends models.PlatformTrends
	json.Unmarshal(trendsJSON, &trends)

	return &trends, true
}

func (m *MemoryCache) Set(platform, keyword string, trends *models.PlatformTrends, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(platform, keyword)

	// Deep copy to avoid mutation
	trendsJSON, _ := json.Marshal(trends)
	var trendsCopy models.PlatformTrends
	json.Unmarshal(trendsJSON, &trendsCopy)

	m.data[key] = cacheEntry{
		trends:    &trendsCopy,
		expiresAt: time.Now().Add(ttl),
	}
}

func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, entry := range m.data {
			if now.After(entry.expiresAt) {
				delete(m.data, key)
			}
		}
		m.mu.Unlock()
	}
}
