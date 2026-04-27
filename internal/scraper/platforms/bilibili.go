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

// BilibiliScraper scrapes Bilibili hot list
type BilibiliScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewBilibiliScraper creates a new Bilibili scraper
func NewBilibiliScraper() *BilibiliScraper {
	return &BilibiliScraper{
		BaseScraper: scraper.NewBaseScraper(
			"bilibili",
			"哔哩哔哩热榜",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type bilibiliResponse struct {
	Code int `json:"code"`
	Data struct {
		List []struct {
			Aid   int64  `json:"aid"`
			Bvid  string `json:"bvid"`
			Title string `json:"title"`
			Desc  string `json:"desc"`
			Pic   string `json:"pic"`
			Stat  struct {
				View int64 `json:"view"`
			} `json:"stat"`
		} `json:"list"`
	} `json:"data"`
}

// Fetch fetches Bilibili hot trends
func (s *BilibiliScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://api.bilibili.com/x/web-interface/ranking/v2?rid=0&type=all"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Referer":    "https://www.bilibili.com/",
		"Origin":     "https://www.bilibili.com",
		"Accept":     "application/json, text/plain, */*",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bilibili: %w", err)
	}

	var resp bilibiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse bilibili response: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("bilibili api error: code=%d", resp.Code)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data.List {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		// Filter by keyword if specified
		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", v.Aid),
			Title:     v.Title,
			Desc:      v.Desc,
			HotValue:  v.Stat.View,
			Pic:       v.Pic,
			URL:       fmt.Sprintf("https://www.bilibili.com/video/%s", v.Bvid),
			MobileURL: fmt.Sprintf("https://m.bilibili.com/video/%s", v.Bvid),
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
