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

// NeteaseScraper scrapes Netease News hot list
type NeteaseScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewNeteaseScraper creates a new Netease scraper
func NewNeteaseScraper() *NeteaseScraper {
	return &NeteaseScraper{
		BaseScraper: scraper.NewBaseScraper(
			"netease",
			"网易新闻",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type neteaseResponse map[string]struct {
	DocID    string `json:"docid"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	ImgSrc   string `json:"imgsrc"`
	Priority int    `json:"priority"`
}

// Fetch fetches Netease hot trends
func (s *NeteaseScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://news.163.com/special/0001386F/rank_whole.html"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch netease: %w", err)
	}

	var resp neteaseResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse netease response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := fmt.Sprintf("https://news.163.com/%s.html", v.DocID)

		items = append(items, models.TrendItem{
			ID:        v.DocID,
			Title:     v.Title,
			Desc:      v.Source,
			HotValue:  int64(v.Priority),
			URL:       itemURL,
			MobileURL: itemURL,
			Pic:       v.ImgSrc,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
