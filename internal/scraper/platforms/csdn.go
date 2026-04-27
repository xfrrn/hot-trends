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

// CSDNScraper scrapes CSDN hot list
type CSDNScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewCSDNScraper creates a new CSDN scraper
func NewCSDNScraper() *CSDNScraper {
	return &CSDNScraper{
		BaseScraper: scraper.NewBaseScraper(
			"csdn",
			"CSDN",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type csdnResponse struct {
	Code int `json:"code"`
	Data []struct {
		ArticleID        string      `json:"articleId"`
		ArticleTitle     string      `json:"articleTitle"`
		CommentCount     interface{} `json:"commentCount"` // Can be string or number
		FavorCount       interface{} `json:"favorCount"`   // Can be string or number
		ArticleDetailURL string      `json:"articleDetailUrl"`
	} `json:"data"`
}

// parseCount converts interface{} to int64, handling both string and number types
func parseCount(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case string:
		var num int64
		fmt.Sscanf(val, "%d", &num)
		return num
	default:
		return 0
	}
}

// Fetch fetches CSDN hot trends
func (s *CSDNScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://blog.csdn.net/phoenix/web/blog/hot-rank?page=0&pageSize=100"

	headers := map[string]string{
		"User-Agent":    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Cache-Control": "no-store",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch csdn: %w", err)
	}

	var resp csdnResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse csdn response: %w", err)
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("csdn api error: code=%d", resp.Code)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		// Filter by keyword if specified
		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.ArticleTitle), strings.ToLower(opts.Keyword)) {
			continue
		}

		items = append(items, models.TrendItem{
			ID:        v.ArticleID,
			Title:     v.ArticleTitle,
			HotValue:  parseCount(v.CommentCount) + parseCount(v.FavorCount),
			URL:       v.ArticleDetailURL,
			MobileURL: v.ArticleDetailURL,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
