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

// WereadScraper scrapes Weread hot list
type WereadScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewWereadScraper creates a new Weread scraper
func NewWereadScraper() *WereadScraper {
	return &WereadScraper{
		BaseScraper: scraper.NewBaseScraper(
			"weread",
			"微信读书",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          10 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type wereadResponse struct {
	Books []struct {
		BookID string `json:"bookId"`
		Title  string `json:"title"`
		Author string `json:"author"`
		Cover  string `json:"cover"`
	} `json:"books"`
}

// Fetch fetches Weread hot trends
func (s *WereadScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://weread.qq.com/web/bookListInCategory/rising?rank=1"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Referer":    "https://weread.qq.com/",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weread: %w", err)
	}

	var resp wereadResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse weread response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for i, v := range resp.Books {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := fmt.Sprintf("https://weread.qq.com/web/reader/%s", v.BookID)

		items = append(items, models.TrendItem{
			ID:        v.BookID,
			Title:     v.Title,
			Desc:      v.Author,
			HotValue:  int64(len(resp.Books) - i),
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
