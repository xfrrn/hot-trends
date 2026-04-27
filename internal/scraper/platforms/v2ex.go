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

// V2exScraper scrapes V2EX hot topics
type V2exScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewV2exScraper creates a new V2EX scraper
func NewV2exScraper() *V2exScraper {
	return &V2exScraper{
		BaseScraper: scraper.NewBaseScraper(
			"v2ex",
			"V2EX",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 30,
				CacheTTL:          10 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type v2exItem struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
	Replies int    `json:"replies"`
}

// Fetch fetches V2EX hot trends
func (s *V2exScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://www.v2ex.com/api/topics/hot.json"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch v2ex: %w", err)
	}

	var resp []v2exItem
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse v2ex response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", v.ID),
			Title:     v.Title,
			Desc:      v.Content,
			HotValue:  int64(v.Replies),
			URL:       v.URL,
			MobileURL: v.URL,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
