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

// XiaohongshuScraper scrapes Xiaohongshu hot list
type XiaohongshuScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewXiaohongshuScraper creates a new Xiaohongshu scraper
func NewXiaohongshuScraper() *XiaohongshuScraper {
	return &XiaohongshuScraper{
		BaseScraper: scraper.NewBaseScraper(
			"xiaohongshu",
			"小红书热榜",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 30, // Lower rate limit due to anti-scraping
				CacheTTL:          10 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type xiaohongshuResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Items []struct {
			ID       string      `json:"id"`
			Title    string      `json:"title"`
			Score    interface{} `json:"score"` // Can be string or number
			WordType string      `json:"word_type"`
		} `json:"items"`
	} `json:"data"`
}

// parseScore converts interface{} to int64, handling both string and number types
func parseScore(v interface{}) int64 {
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

// Fetch fetches Xiaohongshu hot trends
func (s *XiaohongshuScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://edith.xiaohongshu.com/api/sns/v1/search/hot_list"

	// Complex iOS headers for anti-scraping bypass
	headers := map[string]string{
		"User-Agent":       "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.7(0x18000733) NetType/WIFI Language/zh_CN",
		"referer":          "https://app.xhs.cn/",
		"xy-direction":     "22",
		"shield":           "XYAAAAAQAAAAEAAABTAAAAUzUWEe4xG1IYD9/c+qCLOlKGmTtFa+lG434Oe+FTRagxxoaz6rUWSZ3+juJYz8RZqct+oNMyZQxLEBaBEL+H3i0RhOBVGrauzVSARchIWFYwbwkV",
		"xy-platform-info": "platform=iOS&version=8.7&build=8070515&deviceId=C323D3A5-6A27-4CE6-AA0E-51C9D4C26A24&bundle=com.xingin.discover",
		"xy-common-params": "app_id=ECFAAF02&build=8070515&channel=AppStore&deviceId=C323D3A5-6A27-4CE6-AA0E-51C9D4C26A24&device_fingerprint=20230920120211bd7b71a80778509cf4211099ea911000010d2f20f6050264&device_fingerprint1=20230920120211bd7b71a80778509cf4211099ea911000010d2f20f6050264&device_model=phone&fid=1695182528-0-0-63b29d709954a1bb8c8733eb2fb58f29&gid=7dc4f3d168c355f1a886c54a898c6ef21fe7b9a847359afc77fc24ad&identifier_flag=0&lang=zh-Hans&launch_id=716882697&platform=iOS&project_id=ECFAAF&sid=session.1695189743787849952190&t=1695190591&teenager=0&tz=Asia/Shanghai&uis=light&version=8.7",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch xiaohongshu: %w", err)
	}

	var resp xiaohongshuResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse xiaohongshu response: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("xiaohongshu api error: success=false")
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for _, v := range resp.Data.Items {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		// Filter by keyword if specified
		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(opts.Keyword)) {
			continue
		}

		label := ""
		if v.WordType != "" && v.WordType != "无" {
			label = v.WordType
		}

		items = append(items, models.TrendItem{
			ID:        v.ID,
			Title:     v.Title,
			HotValue:  parseScore(v.Score),
			Label:     label,
			URL:       fmt.Sprintf("https://www.xiaohongshu.com/search_result?keyword=%s", url.QueryEscape(v.Title)),
			MobileURL: fmt.Sprintf("https://www.xiaohongshu.com/search_result?keyword=%s", url.QueryEscape(v.Title)),
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
