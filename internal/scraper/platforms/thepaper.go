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

// ThePaperScraper scrapes ThePaper hot list
type ThePaperScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewThePaperScraper creates a new ThePaper scraper
func NewThePaperScraper() *ThePaperScraper {
	return &ThePaperScraper{
		BaseScraper: scraper.NewBaseScraper(
			"thepaper",
			"澎湃新闻",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type thepaperResponse struct {
	Data []struct {
		ContID   string `json:"contid"`
		Title    string `json:"name"`
		Pic      string `json:"pic"`
		PraiseCnt int   `json:"praiseCnt"`
	} `json:"data"`
}

// Fetch fetches ThePaper hot trends
func (s *ThePaperScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://cache.thepaper.cn/contentapi/wwwIndex/rightSidebar"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thepaper: %w", err)
	}

	var resp thepaperResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse thepaper response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := fmt.Sprintf("https://www.thepaper.cn/newsDetail_forward_%s", v.ContID)

		items = append(items, models.TrendItem{
			ID:        v.ContID,
			Title:     v.Title,
			HotValue:  int64(v.PraiseCnt),
			URL:       itemURL,
			MobileURL: itemURL,
			Pic:       v.Pic,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
