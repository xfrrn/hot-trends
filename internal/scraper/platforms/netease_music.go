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

// NeteaseMusicScraper scrapes Netease Cloud Music hot list
type NeteaseMusicScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewNeteaseMusicScraper creates a new Netease Music scraper
func NewNeteaseMusicScraper() *NeteaseMusicScraper {
	return &NeteaseMusicScraper{
		BaseScraper: scraper.NewBaseScraper(
			"netease_music",
			"网易云音乐",
			scraper.JSONAPIScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 60,
				CacheTTL:          10 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

type neteaseMusicResponse struct {
	Playlist struct {
		Tracks []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Ar   []struct {
				Name string `json:"name"`
			} `json:"ar"`
			Al struct {
				PicURL string `json:"picUrl"`
			} `json:"al"`
		} `json:"tracks"`
	} `json:"playlist"`
}

// Fetch fetches Netease Music hot trends
func (s *NeteaseMusicScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://music.163.com/api/playlist/detail?id=3778678"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Referer":    "https://music.163.com/",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch netease music: %w", err)
	}

	var resp neteaseMusicResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse netease music response: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)
	for i, v := range resp.Playlist.Tracks {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			break
		}

		artist := ""
		if len(v.Ar) > 0 {
			artist = v.Ar[0].Name
		}

		if opts.Keyword != "" && !strings.Contains(strings.ToLower(v.Name), strings.ToLower(opts.Keyword)) {
			continue
		}

		itemURL := fmt.Sprintf("https://music.163.com/#/song?id=%d", v.ID)

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", v.ID),
			Title:     v.Name,
			Desc:      artist,
			HotValue:  int64(len(resp.Playlist.Tracks) - i),
			URL:       itemURL,
			MobileURL: itemURL,
			Pic:       v.Al.PicURL,
		})
	}

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
