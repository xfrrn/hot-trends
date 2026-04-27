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

// HupuScraper scrapes Hupu hot list
type HupuScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewHupuScraper creates a new Hupu scraper
func NewHupuScraper() *HupuScraper {
	return &HupuScraper{
		BaseScraper: scraper.NewBaseScraper(
			"hupu",
			"虎扑",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type hupuResponse struct {
	Result []struct {
		Tid       int    `json:"tid"`
		Title     string `json:"title"`
		Replies   int    `json:"replies"`
		URL       string `json:"url"`
	} `json:"result"`
}

// Fetch fetches Hupu hot trends
func (s *HupuScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://games.mobileapi.hupu.com/1/7.5.60/bbs/hotThread"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hupu: %w", err)
	}

	var resp hupuResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse hupu response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Result {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := v.URL
		if itemURL == "" {
			itemURL = fmt.Sprintf("https://bbs.hupu.com/%d.html", v.Tid)
		}

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", v.Tid),
			Title:     v.Title,
			HotValue:  int64(v.Replies),
			URL:       itemURL,
			MobileURL: itemURL,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
