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

// JuejinScraper scrapes Juejin hot list
type JuejinScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewJuejinScraper creates a new Juejin scraper
func NewJuejinScraper() *JuejinScraper {
	return &JuejinScraper{
		BaseScraper: scraper.NewBaseScraper(
			"juejin",
			"稀土掘金",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type juejinResponse struct {
	ErrMsg string `json:"err_msg"`
	Data   []struct {
		Content struct {
			ArticleID string `json:"article_id"`
			Title     string `json:"title"`
			BriefContent string `json:"brief_content"`
		} `json:"content"`
		ContentCounter struct {
			View int64 `json:"view"`
		} `json:"content_counter"`
	} `json:"data"`
}

// Fetch fetches Juejin hot trends
func (s *JuejinScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://api.juejin.cn/content_api/v1/content/article_rank?category_id=1&type=hot"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch juejin: %w", err)
	}

	var resp juejinResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse juejin response: %w", err)
	}

	if resp.ErrMsg != "success" {
		return nil, fmt.Errorf("juejin api error: err_msg=%s", resp.ErrMsg)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		// Filter by keyword if specified
		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Content.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        v.Content.ArticleID,
			Title:     v.Content.Title,
			Desc:      v.Content.BriefContent,
			HotValue:  v.ContentCounter.View,
			URL:       fmt.Sprintf("https://juejin.cn/post/%s", v.Content.ArticleID),
			MobileURL: fmt.Sprintf("https://juejin.cn/post/%s", v.Content.ArticleID),
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
