package platforms

import (
	"context"
	"encoding/json"
	"fmt"
	"hot-trends-service/internal/models"
	"hot-trends-service/internal/scraper"
	"hot-trends-service/pkg/httpclient"
	"net/url"
	"strings"
	"time"
)

// WeiboScraper scrapes Weibo hot search
type WeiboScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewWeiboScraper creates a new Weibo scraper
func NewWeiboScraper() *WeiboScraper {
	return &WeiboScraper{
		BaseScraper: scraper.NewBaseScraper(
			"weibo",
			"微博热搜",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type weiboResponse struct {
	Ok   int `json:"ok"`
	Data struct {
		Realtime []struct {
			Mid        string `json:"mid"`
			Word       string `json:"word"`
			WordScheme string `json:"word_scheme"`
			Num        int64  `json:"num"`
			LabelName  string `json:"label_name"`
		} `json:"realtime"`
	} `json:"data"`
}

// Fetch fetches Weibo hot trends
func (s *WeiboScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://weibo.com/ajax/side/hotSearch"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Referer":    "https://weibo.com/",
		"Accept":     "application/json",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weibo: %w", err)
	}

	var resp weiboResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse weibo response: %w", err)
	}

	if resp.Ok != 1 {
		return nil, fmt.Errorf("weibo api error: ok=%d", resp.Ok)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data.Realtime {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		// Filter by keyword if specified
		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Word), strings.ToLower(opts.Keyword)) {
			continue
		}

		key := v.WordScheme
		if key == "" {
			key = "#" + v.Word + "#"
		}

		// Map label names to emojis
		label := ""
		switch v.LabelName {
		case "热":
			label = "🔥"
		case "沸":
			label = "💥"
		case "新":
			label = "✨"
		case "暖":
			label = "💛"
		case "爆":
			label = "💣"
		}

		items = append(items, models.TrendItem{
			ID:        v.Mid,
			Title:     v.Word,
			Desc:      key,
			HotValue:  v.Num,
			URL:       fmt.Sprintf("https://s.weibo.com/weibo?q=%s&t=31&band_rank=1&Refer=top", url.QueryEscape(key)),
			MobileURL: fmt.Sprintf("https://s.weibo.com/weibo?q=%s&t=31&band_rank=1&Refer=top", url.QueryEscape(key)),
			Label:     label,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
