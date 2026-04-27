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

// ToutiaoScraper scrapes Toutiao hot list
type ToutiaoScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewToutiaoScraper creates a new Toutiao scraper
func NewToutiaoScraper() *ToutiaoScraper {
	return &ToutiaoScraper{
		BaseScraper: scraper.NewBaseScraper(
			"toutiao",
			"今日头条",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type toutiaoResponse struct {
	Status string `json:"status"`
	Data   []struct {
		ClusterIDStr string `json:"ClusterIdStr"`
		Title        string `json:"Title"`
		Image        struct {
			URL string `json:"url"`
		} `json:"Image"`
		HotValue interface{} `json:"HotValue"` // Can be string or number
		URL      string      `json:"Url"`
	} `json:"data"`
}

// parseHotValue converts interface{} to int64, handling both string and number types
func parseHotValue(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case string:
		var num int64
		fmt.Sscanf(val, "%d", &num)
		return num
	default:
		return 0
	}
}

// Fetch fetches Toutiao hot trends
func (s *ToutiaoScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://www.toutiao.com/hot-event/hot-board/?origin=toutiao_pc"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch toutiao: %w", err)
	}

	var resp toutiaoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse toutiao response: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("toutiao api error: status=%s", resp.Status)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		// Filter by keyword if specified
		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        v.ClusterIDStr,
			Title:     v.Title,
			HotValue:  parseHotValue(v.HotValue),
			Pic:       v.Image.URL,
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
