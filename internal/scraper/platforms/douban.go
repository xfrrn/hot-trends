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

// DoubanScraper scrapes Douban movie hot list
type DoubanScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewDoubanScraper creates a new Douban scraper
func NewDoubanScraper() *DoubanScraper {
	return &DoubanScraper{
		BaseScraper: scraper.NewBaseScraper(
			"douban",
			"豆瓣电影",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 30,
				CacheTTL:          10 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type doubanResponse struct {
	Subjects []struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Rate   string `json:"rate"`
		Cover  string `json:"cover"`
		URL    string `json:"url"`
	} `json:"subjects"`
}

// Fetch fetches Douban hot trends
func (s *DoubanScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://movie.douban.com/j/search_subjects?type=movie&tag=%E7%83%AD%E9%97%A8&page_limit=50&page_start=0"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Referer":    "https://movie.douban.com/",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch douban: %w", err)
	}

	var resp doubanResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse douban response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for i, v := range resp.Subjects {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        v.ID,
			Title:     v.Title,
			Desc:      "评分: " + v.Rate,
			HotValue:  int64(len(resp.Subjects) - i),
			URL:       v.URL,
			MobileURL: v.URL,
			Pic:       v.Cover,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
