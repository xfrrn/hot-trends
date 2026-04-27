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

// Kr36Scraper scrapes 36Kr hot list
type Kr36Scraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewKr36Scraper creates a new 36Kr scraper
func NewKr36Scraper() *Kr36Scraper {
	return &Kr36Scraper{
		BaseScraper: scraper.NewBaseScraper(
			"kr36",
			"36氪",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type kr36Response struct {
	Data struct {
		HotList []struct {
			ID             int    `json:"id"`
			TemplateID     int    `json:"templateMaterial"`
			ItemID         int    `json:"itemId"`
			Title          string `json:"title"`
			Summary        string `json:"summary"`
			Cover          string `json:"cover"`
			NewsFlashID    int    `json:"newsflashId"`
			TemplateType   string `json:"templateType"`
		} `json:"hotRankList"`
	} `json:"data"`
}

// Fetch fetches 36Kr hot trends
func (s *Kr36Scraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://gateway.36kr.com/api/mis/nav/home/nav/rank/hot"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch 36kr: %w", err)
	}

	var resp kr36Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse 36kr response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for i, v := range resp.Data.HotList {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := ""
		if v.TemplateID == 1 && v.ItemID > 0 {
			itemURL = fmt.Sprintf("https://www.36kr.com/p/%d", v.ItemID)
		} else if v.NewsFlashID > 0 {
			itemURL = fmt.Sprintf("https://www.36kr.com/newsflashes/%d", v.NewsFlashID)
		}

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", v.ID),
			Title:     v.Title,
			Desc:      v.Summary,
			HotValue:  int64(len(resp.Data.HotList) - i),
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
