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

// NgabbsScraper scrapes NGA BBS hot topics
type NgabbsScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewNgabbsScraper creates a new NGA scraper
func NewNgabbsScraper() *NgabbsScraper {
	return &NgabbsScraper{
		BaseScraper: scraper.NewBaseScraper(
			"ngabbs",
			"NGA论坛",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type ngabbsResponse struct {
	Result []struct {
		TID     int    `json:"tid"`
		Subject string `json:"subject"`
		Author  string `json:"author"`
		Replies int    `json:"replies"`
	} `json:"result"`
}

// Fetch fetches NGA hot trends
func (s *NgabbsScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://ngabbs.com/nuke.php?func=hottopic&__output=8"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ngabbs: %w", err)
	}

	var resp ngabbsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse ngabbs response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Result {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Subject), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := fmt.Sprintf("https://ngabbs.com/read.php?tid=%d", v.TID)

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", v.TID),
			Title:     v.Subject,
			Desc:      v.Author,
			HotValue:  int64(v.Replies),
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
