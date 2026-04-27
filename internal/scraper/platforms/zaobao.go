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

// ZaobaoChinaScraper scrapes Zaobao China news
type ZaobaoChinaScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewZaobaoChinaScraper creates a new Zaobao China scraper
func NewZaobaoChinaScraper() *ZaobaoChinaScraper {
	return &ZaobaoChinaScraper{
		BaseScraper: scraper.NewBaseScraper(
			"zaobao",
			"联合早报",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          10 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type zaobaoResponse struct {
	Result struct {
		Items []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Desc  string `json:"description"`
			Link  string `json:"link"`
		} `json:"items"`
	} `json:"result"`
}

// Fetch fetches Zaobao hot trends
func (s *ZaobaoChinaScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://www.zaobao.com/wencui/social"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch zaobao: %w", err)
	}

	var resp zaobaoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse zaobao response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for i, v := range resp.Result.Items {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        v.ID,
			Title:     v.Title,
			Desc:      v.Desc,
			HotValue:  int64(len(resp.Result.Items) - i),
			URL:       v.Link,
			MobileURL: v.Link,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
