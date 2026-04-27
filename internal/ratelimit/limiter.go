package ratelimit

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

// PlatformLimiter manages rate limiters for each platform
type PlatformLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewPlatformLimiter creates a new platform rate limiter
func NewPlatformLimiter() *PlatformLimiter {
	return &PlatformLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// Register registers a rate limiter for a platform
func (pl *PlatformLimiter) Register(platform string, requestsPerMinute int) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// Convert requests per minute to rate.Limit
	// e.g., 60 requests/min = 1 request/second
	limiter := rate.NewLimiter(rate.Limit(float64(requestsPerMinute)/60.0), requestsPerMinute)
	pl.limiters[platform] = limiter
}

// Wait waits for permission to make a request to the platform
func (pl *PlatformLimiter) Wait(ctx context.Context, platform string) error {
	pl.mu.RLock()
	limiter, ok := pl.limiters[platform]
	pl.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no rate limiter for platform: %s", platform)
	}

	return limiter.Wait(ctx)
}
