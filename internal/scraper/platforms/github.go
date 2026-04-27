package platforms

import (
	"context"
	"fmt"
	"hot-trends-service/internal/models"
	"hot-trends-service/internal/scraper"
	"hot-trends-service/pkg/httpclient"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// GitHubScraper scrapes GitHub trending
type GitHubScraper struct {
	*scraper.BaseScraper
	client *httpclient.Client
}

// NewGitHubScraper creates a new GitHub scraper
func NewGitHubScraper() *GitHubScraper {
	return &GitHubScraper{
		BaseScraper: scraper.NewBaseScraper(
			"github",
			"GitHub Trending",
			scraper.HTMLScraper,
			scraper.RateLimitConfig{
				RequestsPerMinute: 30,
				CacheTTL:          10 * time.Minute,
			},
		),
		client: httpclient.NewClient(10*time.Second, 3),
	}
}

// Fetch fetches GitHub trending repositories
func (s *GitHubScraper) Fetch(ctx context.Context, opts scraper.FetchOptions) (*models.PlatformTrends, error) {
	apiURL := "https://github.com/trending"

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	body, err := s.client.GetWithRetry(ctx, apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch github: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse github html: %w", err)
	}

	items := make([]models.TrendItem, 0, opts.Limit)

	doc.Find("article.Box-row").Each(func(i int, sel *goquery.Selection) {
		if opts.Limit > 0 && len(items) >= opts.Limit {
			return
		}

		// Extract repository name
		repoLink := sel.Find("h2 a")
		repoPath, _ := repoLink.Attr("href")
		repoName := strings.TrimPrefix(repoPath, "/")

		// Filter by keyword if specified
		if opts.Keyword != "" && !strings.Contains(strings.ToLower(repoName), strings.ToLower(opts.Keyword)) {
			return
		}

		// Extract description
		desc := strings.TrimSpace(sel.Find("p.col-9").Text())

		// Extract stars today
		starsToday := strings.TrimSpace(sel.Find("span.d-inline-block.float-sm-right").Text())

		items = append(items, models.TrendItem{
			ID:        fmt.Sprintf("%d", i+1),
			Title:     repoName,
			Desc:      desc,
			Label:     starsToday,
			URL:       "https://github.com" + repoPath,
			MobileURL: "https://github.com" + repoPath,
		})
	})

	return &models.PlatformTrends{
		Platform:  s.Name(),
		Items:     items,
		FetchedAt: time.Now(),
	}, nil
}
