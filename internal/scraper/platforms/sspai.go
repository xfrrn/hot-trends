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

// SspaiScraper scrapes Sspai hot list
type SspaiScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewSspaiScraper creates a new Sspai scraper
func NewSspaiScraper() *SspaiScraper {
	return &SspaiScraper{
		BaseScraper: scraper.NewBaseScraper(
			"sspai",
			"少数派",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          10 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type sspaiResponse struct {
	Data []struct {
		ID       int    `json:"id"`
		Title    string `json:"title"`
		Summary  string `json:"summary"`
		Banner   string `json:"banner"`
		LikeCount int   `json:"like_count"`
	} `json:"data"`
}

// Fetch fetches Sspai hot trends
func (s *SspaiScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://sspai.com/api/v1/article/tag/page/get?limit=50&offset=0&sort=hot&tag=%E7%83%AD%E9%97%A8"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sspai: %w", err)
	}

	var resp sspaiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse sspai response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := fmt.Sprintf("https://sspai.com/post/%d", v.ID)

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", v.ID),
			Title:     v.Title,
			Desc:      v.Summary,
			HotValue:  int64(v.LikeCount),
			URL:       itemURL,
			MobileURL: itemURL,
			Pic:       v.Banner,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
