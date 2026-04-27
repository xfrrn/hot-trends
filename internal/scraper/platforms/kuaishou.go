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

// KuaishouScraper scrapes Kuaishou hot list
type KuaishouScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewKuaishouScraper creates a new Kuaishou scraper
func NewKuaishouScraper() *KuaishouScraper {
	return &KuaishouScraper{
		BaseScraper: scraper.NewBaseScraper(
			"kuaishou",
			"快手",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type kuaishouResponse struct {
	Data struct {
		Items []struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Cover    string `json:"cover"`
			ViewCount int64 `json:"viewCount"`
		} `json:"items"`
	} `json:"data"`
}

// Fetch fetches Kuaishou hot trends
func (s *KuaishouScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://www.kuaishou.com/graphql"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Content-Type": "application/json",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch kuaishou: %w", err)
	}

	var resp kuaishouResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse kuaishou response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data.Items {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := fmt.Sprintf("https://www.kuaishou.com/short-video/%s", v.ID)

		items = append(items, models.TrendItem{
			ID:        v.ID,
			Title:     v.Title,
			HotValue:  v.ViewCount,
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
