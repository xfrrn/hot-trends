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

// TiebaScraper scrapes Baidu Tieba hot topics
type TiebaScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewTiebaScraper creates a new Tieba scraper
func NewTiebaScraper() *TiebaScraper {
	return &TiebaScraper{
		BaseScraper: scraper.NewBaseScraper(
			"tieba",
			"百度贴吧",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type tiebaResponse struct {
	Data struct {
		Bang struct {
			ThreadList []struct {
				ThreadID   string `json:"thread_id"`
				Title      string `json:"title"`
				Desc       string `json:"desc"`
				ThreadName string `json:"thread_name"`
				HotScore   int64  `json:"hot_score"`
			} `json:"thread_list"`
		} `json:"bang"`
	} `json:"data"`
}

// Fetch fetches Tieba hot trends
func (s *TiebaScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://tieba.baidu.com/hottopic/browse/topicList"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tieba: %w", err)
	}

	var resp tiebaResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse tieba response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data.Bang.ThreadList {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := fmt.Sprintf("https://tieba.baidu.com/p/%s", v.ThreadID)

		items = append(items, models.TrendItem{
			ID:        v.ThreadID,
			Title:     v.Title,
			Desc:      v.Desc,
			HotValue:  v.HotScore,
			URL:       itemURL,
			MobileURL: itemURL,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
