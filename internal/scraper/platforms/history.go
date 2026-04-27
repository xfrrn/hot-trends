package platforms

import (
	"context"
	"encoding/json"
	"fmt"
	"hot-trends-service/internal/models"
	"hot-trends-service/internal/scraper"
	"hot-trends-service/pkg/httpclient"
	"strings"
	"time"
)

// HistoryScraper scrapes history events
type HistoryScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewHistoryScraper creates a new History scraper
func NewHistoryScraper() *HistoryScraper {
	return &HistoryScraper{
		BaseScraper: scraper.NewBaseScraper(
			"history",
			"历史上的今天",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          60 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type historyResponse struct {
	Result []struct {
		Year  string `json:"year"`
		Title string `json:"title"`
		Desc  string `json:"desc"`
		Cover string `json:"cover"`
	} `json:"result"`
}

// Fetch fetches history events
func (s *HistoryScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	now := time.Now()
	apiURL := fmt.Sprintf("https://baike.baidu.com/cms/home/eventsOnHistory/%02d.json", now.Month())

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch history: %w", err)
	}

	var resp historyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse history response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for i, v := range resp.Result {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", i),
			Title:     fmt.Sprintf("%s年 - %s", v.Year, v.Title),
			Desc:      v.Desc,
			HotValue:  int64(len(resp.Result) - i),
			URL:       "https://baike.baidu.com/calendar/",
			MobileURL: "https://baike.baidu.com/calendar/",
			Pic:       v.Cover,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
