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

// BaiduScraper scrapes Baidu hot search
type BaiduScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewBaiduScraper creates a new Baidu scraper
func NewBaiduScraper() *BaiduScraper {
	return &BaiduScraper{
		BaseScraper: scraper.NewBaseScraper(
			"baidu",
			"百度热搜",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type baiduResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Cards []struct {
			Content []struct {
				Content []struct {
					Word  string `json:"word"`
					Query string `json:"query"`
					Desc  string `json:"desc"`
					HotScore string `json:"hotScore"`
					Img   string `json:"img"`
					URL   string `json:"url"`
				} `json:"content"`
			} `json:"content"`
		} `json:"cards"`
	} `json:"data"`
}

// Fetch fetches Baidu hot trends
func (s *BaiduScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://top.baidu.com/api/board?platform=wise&tab=realtime"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch baidu: %w", err)
	}

	var resp baiduResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse baidu response: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("baidu api error: success=false")
	}

	items := make([]models.TrendItem, 0, opts.Limit)

	if len(resp.Data.Cards) > 0 && len(resp.Data.Cards[0].Content) > 0 {
		for i, v := range resp.Data.Cards[0].Content[0].Content {
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
				Desc:      v.Desc,
				Pic:       v.Img,
				Label:     v.HotScore,
				URL:       v.URL,
				MobileURL: v.URL,
			})
		}
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
