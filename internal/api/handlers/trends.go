package handlers

import (
	"fmt"
	"hot-trends-service/internal/models"
	"hot-trends-service/internal/scraper"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TrendsHandler handles trends requests
type TrendsHandler struct {
	executor *scraper.Executor
	registry *scraper.Registry
}

// NewTrendsHandler creates a new trends handler
func NewTrendsHandler(executor *scraper.Executor, registry *scraper.Registry) *TrendsHandler {
	return &TrendsHandler{
		executor: executor,
		registry: registry,
	}
}

// GetSingle handles single platform trends request
func (h *TrendsHandler) GetSingle(c *gin.Context) {
	platform := c.Param("platform")
	limit := 10
	if l, ok := c.GetQuery("limit"); ok {
		if parsed, err := parseLimit(l); err == nil {
			limit = parsed
		}
	}

	keyword := c.Query("keyword")

	opts := scraper.FetchOptions{
		Limit:   limit,
		Keyword: keyword,
		Timeout: 10 * time.Second,
	}

	result := h.executor.FetchSingle(c.Request.Context(), platform, opts)

	c.JSON(http.StatusOK, result)
}

// GetBatch handles batch trends request
func (h *TrendsHandler) GetBatch(c *gin.Context) {
	var req models.BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	timeout := 30 * time.Second
	if req.TimeoutSeconds > 0 {
		timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}

	opts := scraper.FetchOptions{
		Limit:   req.Limit,
		Keyword: req.Keyword,
		Timeout: timeout,
	}

	results := h.executor.FetchMultiple(c.Request.Context(), req.Platforms, opts)

	successful := 0
	failed := 0
	for _, r := range results {
		if r.Error == "" {
			successful++
		} else {
			failed++
		}
	}

	c.JSON(http.StatusOK, models.BatchResponse{
		Results:        results,
		TotalPlatforms: len(req.Platforms),
		Successful:     successful,
		Failed:         failed,
	})
}

// ListPlatforms lists all available platforms
func (h *TrendsHandler) ListPlatforms(c *gin.Context) {
	scrapers := h.registry.List()
	platforms := make([]models.PlatformInfo, 0, len(scrapers))

	for _, s := range scrapers {
		platforms = append(platforms, models.PlatformInfo{
			ID:        s.Name(),
			Name:      s.DisplayName(),
			Type:      string(s.Type()),
			RateLimit: formatRateLimit(s.RateLimit().RequestsPerMinute),
			CacheTTL:  s.RateLimit().CacheTTL.String(),
			Enabled:   true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"platforms": platforms,
	})
}

func parseLimit(s string) (int, error) {
	var limit int
	_, err := fmt.Sscanf(s, "%d", &limit)
	if err != nil {
		return 0, err
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return limit, nil
}

func formatRateLimit(rpm int) string {
	return fmt.Sprintf("%d/min", rpm)
}
