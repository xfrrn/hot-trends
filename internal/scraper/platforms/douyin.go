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

// DouyinScraper scrapes Douyin hot search
type DouyinScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewDouyinScraper creates a new Douyin scraper
func NewDouyinScraper() *DouyinScraper {
	return &DouyinScraper{
		BaseScraper: scraper.NewBaseScraper(
			"douyin",
			"抖音热搜",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type douyinResponse struct {
	StatusCode int `json:"status_code"`
	Data       struct {
		WordList []struct {
			Word      string `json:"word"`
			HotValue  int64  `json:"hot_value"`
			SentenceID string `json:"sentence_id"`
			EventTime int64  `json:"event_time"`
		} `json:"word_list"`
	} `json:"data"`
}

// Fetch fetches Douyin hot trends
func (s *DouyinScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://aweme.snssdk.com/aweme/v1/hot/search/list/"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch douyin: %w", err)
	}

	var resp douyinResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse douyin response: %w", err)
	}

	if resp.StatusCode != 0 {
		return nil, fmt.Errorf("douyin api error: status_code=%d", resp.StatusCode)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for i, v := range resp.Data.WordList {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		// Filter by keyword if specified
		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Word), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", i+1),
			Title:     v.Word,
			HotValue:  v.HotValue,
			URL:       fmt.Sprintf("https://www.douyin.com/search/%s", v.Word),
			MobileURL: fmt.Sprintf("https://www.douyin.com/search/%s", v.Word),
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
