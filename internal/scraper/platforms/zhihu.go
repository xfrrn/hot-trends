package platforms

import (
	"context"
	"encoding/json"
	"fmt"
	"hot-trends-service/internal/models"
	"hot-trends-service/internal/scraper"
	"hot-trends-service/pkg/httpclient"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ZhihuScraper scrapes Zhihu hot list
type ZhihuScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewZhihuScraper creates a new Zhihu scraper
func NewZhihuScraper() *ZhihuScraper {
	return &ZhihuScraper{
		BaseScraper: scraper.NewBaseScraper(
			"zhihu",
			"知乎热榜",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          5 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type zhihuResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Target     struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			URL   string `json:"url"`
			Excerpt string `json:"excerpt"`
		} `json:"target"`
		DetailText string `json:"detail_text"`
	} `json:"data"`
}

// Fetch fetches Zhihu hot trends
func (s *ZhihuScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://api.zhihu.com/topstory/hot-list"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch zhihu: %w", err)
	}

	var resp zhihuResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse zhihu response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		// Filter by keyword if specified
		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Target.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		// Parse hot value from detail_text (e.g., "123万热度" -> 1230000)
		hotValue := parseZhihuHotValue(v.DetailText)

		items = append(items, models.TrendItem{
			ID:        v.ID,
			Title:     v.Target.Title,
			Desc:      v.Target.Excerpt,
			HotValue:  hotValue,
			URL:       fmt.Sprintf("https://www.zhihu.com/question/%d", v.Target.ID),
			MobileURL: fmt.Sprintf("https://www.zhihu.com/question/%d", v.Target.ID),
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}

// parseZhihuHotValue parses hot value from text like "123万热度" or "1234热度"
func parseZhihuHotValue(text string) int64 {
	// Remove "热度" suffix
	text = strings.TrimSuffix(text, "热度")
	text = strings.TrimSpace(text)

	// Check if it contains "万" (10,000)
	if strings.Contains(text, "万") {
		text = strings.TrimSuffix(text, "万")
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			return int64(val * 10000)
		}
	}

	// Try to parse as integer
	re := regexp.MustCompile(`\d+`)
	matches := re.FindString(text)
	if matches != "" {
		if val, err := strconv.ParseInt(matches, 10, 64); err == nil {
			return val
		}
	}

	return 0
}
