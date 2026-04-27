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

// HuxiuScraper scrapes Huxiu hot list
type HuxiuScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewHuxiuScraper creates a new Huxiu scraper
func NewHuxiuScraper() *HuxiuScraper {
	return &HuxiuScraper{
		BaseScraper: scraper.NewBaseScraper(
			"huxiu",
			"虎嗅",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          10 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type huxiuResponse struct {
	Data struct {
		DataList []struct {
			AID       int    `json:"aid"`
			Title     string `json:"title"`
			Desc      string `json:"desc"`
			Cover     string `json:"cover"`
			FavorCnt  int    `json:"favor_count"`
		} `json:"dataList"`
	} `json:"data"`
}

// Fetch fetches Huxiu hot trends
func (s *HuxiuScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://www.huxiu.com/v2/article/list.json"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch huxiu: %w", err)
	}

	var resp huxiuResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse huxiu response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data.DataList {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := fmt.Sprintf("https://www.huxiu.com/article/%d.html", v.AID)

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", v.AID),
			Title:     v.Title,
			Desc:      v.Desc,
			HotValue:  int64(v.FavorCnt),
			URL:       itemURL,
			MobileURL: itemURL,
			Pic:       v.Cover,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
