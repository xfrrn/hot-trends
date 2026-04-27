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

// ItHomeScraper scrapes IT Home news
type ItHomeScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewItHomeScraper creates a new IT Home scraper
func NewItHomeScraper() *ItHomeScraper {
	return &ItHomeScraper{
		BaseScraper: scraper.NewBaseScraper(
			"ithome",
			"IT之家",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type ithomeResponse struct {
	Newslist []struct {
		NewsID       int    `json:"newsid"`
		Title        string `json:"title"`
		URL          string `json:"url"`
		Image        string `json:"image"`
		CommentCount int    `json:"commentcount"`
	} `json:"newslist"`
}

// Fetch fetches IT Home hot trends
func (s *ItHomeScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://m.ithome.com/api/news/newslistpageget"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ithome: %w", err)
	}

	var resp ithomeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse ithome response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Newslist {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", v.NewsID),
			Title:     v.Title,
			HotValue:  int64(v.CommentCount),
			URL:       v.URL,
			MobileURL: v.URL,
			Pic:       v.Image,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
