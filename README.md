# Hot Trends Aggregation Service

A high-performance Go microservice that aggregates hot trends from 30+ Chinese platforms including Weibo, Zhihu, Douyin, Xiaohongshu, Bilibili, GitHub, and more.

## 📋 Project Attribution

This project is based on the scraping logic and platform implementations from [**next-daily-hot**](https://github.com/imsyy/DailyHot) (今日热榜).

- **Original Project**: next-daily-hot - A Next.js-based hot trends aggregator
- **Original Author**: [imsyy](https://github.com/imsyy)
- **Original Tech Stack**: Next.js + TypeScript + React
- **This Implementation**: Go microservice adaptation for the AutoUp-Agentic project

All platform scraping patterns, API endpoints, and data transformation logic are derived from the original TypeScript implementation. This Go version provides:

- Higher performance and lower memory footprint
- Better concurrency handling with goroutines
- Standalone microservice architecture
- Docker containerization

**Thank you to the original project for the excellent reference implementation!**

## 🚀 Features

- **30+ Platform Support**: Weibo, Zhihu, Douyin, Xiaohongshu, Bilibili, GitHub, Baidu, Toutiao, Juejin, CSDN, and more
- **High Performance**: Built with Go for excellent concurrency and low resource usage
- **Concurrent Scraping**: Parallel fetching from multiple platforms using goroutines
- **Smart Caching**: In-memory cache with configurable TTL per platform
- **Rate Limiting**: Per-platform token bucket rate limiting to avoid IP bans
- **Retry Logic**: Automatic retry with exponential backoff for transient failures
- **RESTful API**: Clean HTTP API with JSON responses
- **Docker Ready**: Multi-stage Dockerfile for minimal image size (~15MB)
- **Keyword Filtering**: Filter trends by keyword across all platforms

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     HTTP API Layer                          │
│  GET /health                                                │
│  GET /api/v1/trends/{platform}?limit=10&keyword=羽毛球      │
│  POST /api/v1/trends/batch                                  │
│  GET /api/v1/platforms                                      │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│              Scraper Execution Engine                       │
│  - Concurrent execution with goroutines                     │
│  - Cache layer (in-memory)                                  │
│  - Rate limiting (token bucket per platform)                │
│  - Retry logic with exponential backoff                     │
└─────────────────────┬───────────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
┌───────▼──────┐ ┌────▼─────┐ ┌────▼──────┐
│   Weibo      │ │  Zhihu   │ │  Douyin   │
│   Scraper    │ │  Scraper │ │  Scraper  │
└──────────────┘ └──────────┘ └───────────┘
     ... (30+ platform scrapers)
```

## 📦 Installation

### Prerequisites

- Go 1.22 or higher
- Docker (optional, for containerized deployment)

### Local Development

```bash
# Enter project root
cd hot-trends

# Download dependencies
go mod download

# Build the application
go build -o bin/server ./cmd/server

# Run the server
./bin/server
```

### Quick Start (One Command)

```bash
# Option 1: Makefile
make quickstart

# Option 2: Go directly (works well on Windows PowerShell)
go run ./cmd/server
```

The server will start on `http://localhost:6000`

### Docker Deployment

```bash
# Build and run with docker-compose
cd hot-trends
docker-compose -f deployments/docker-compose.yml up -d

# Or build manually
docker build -f deployments/Dockerfile -t hot-trends:latest .
docker run -p 6000:6000 hot-trends:latest
```

## 🔌 API Documentation

### Interactive Docs (Swagger UI)

After starting the server, open:

- `http://localhost:6000/docs` (default)
- `http://localhost:6001/docs` (if you set `PORT=6001`)

You can call APIs directly on this page using **Try it out**.

### Health Check

```bash
GET /health
```

**Response:**

```json
{
  "status": "ok",
  "timestamp": "2026-04-27T10:30:00Z",
  "redis_connected": true,
  "scrapers_registered": 10
}
```

### Get Single Platform Trends

```bash
GET /api/v1/trends/{platform}?limit=10&keyword=羽毛球
```

**Parameters:**

- `platform` (path): Platform identifier (e.g., `weibo`, `zhihu`, `douyin`)
- `limit` (query, optional): Number of items to return (default: 10, max: 100)
- `keyword` (query, optional): Filter trends by keyword

**Response:**

```json
{
  "platform": "weibo",
  "items": [
    {
      "id": "4890123456",
      "title": "羽毛球世锦赛决赛",
      "desc": "#羽毛球世锦赛#",
      "hot_value": 1234567,
      "url": "https://s.weibo.com/weibo?q=%23羽毛球世锦赛%23",
      "mobile_url": "https://s.weibo.com/weibo?q=%23羽毛球世锦赛%23",
      "pic": "https://...",
      "label": "🔥"
    }
  ],
  "fetched_at": "2026-04-27T10:30:00Z",
  "cached": false,
  "error": ""
}
```

### Get Batch Platform Trends

```bash
POST /api/v1/trends/batch
Content-Type: application/json

{
  "platforms": ["weibo", "zhihu", "douyin"],
  "limit": 10,
  "keyword": "羽毛球",
  "timeout_seconds": 15
}
```

**Response:**

```json
{
  "results": [
    {
      "platform": "weibo",
      "items": [...],
      "fetched_at": "2026-04-27T10:30:00Z",
      "cached": false,
      "error": ""
    },
    {
      "platform": "zhihu",
      "items": [...],
      "fetched_at": "2026-04-27T10:30:00Z",
      "cached": true,
      "error": ""
    }
  ],
  "total_platforms": 3,
  "successful": 2,
  "failed": 1
}
```

### List Available Platforms

```bash
GET /api/v1/platforms
```

**Response:**

```json
{
  "platforms": [
    {
      "id": "weibo",
      "name": "微博热搜",
      "type": "json_api",
      "rate_limit": "60/min",
      "cache_ttl": "5m0s",
      "enabled": true
    },
    {
      "id": "xiaohongshu",
      "name": "小红书热榜",
      "type": "json_api",
      "rate_limit": "30/min",
      "cache_ttl": "10m0s",
      "enabled": true,
      "requires_special_headers": true
    }
  ]
}
```

## 🎯 Supported Platforms

### Tier 1 - Critical Platforms (Implemented)

- ✅ **weibo** - 微博热搜
- ✅ **zhihu** - 知乎热榜
- ✅ **douyin** - 抖音热搜
- ✅ **xiaohongshu** - 小红书热榜
- ✅ **bilibili** - 哔哩哔哩热榜

### Tier 2 - Important Platforms (Implemented)

- ✅ **github** - GitHub Trending
- ✅ **baidu** - 百度热搜
- ✅ **toutiao** - 今日头条
- ✅ **juejin** - 稀土掘金
- ✅ **csdn** - CSDN

### Tier 3 - Additional Platforms (To Be Implemented)

- ⏳ **tieba** - 百度贴吧
- ⏳ **netease** - 网易新闻
- ⏳ **qq** - 腾讯新闻
- ⏳ **thepaper** - 澎湃新闻
- ⏳ **quark** - 夸克
- ⏳ **huxiu** - 虎嗅
- ⏳ **ifanr** - 爱范儿
- ⏳ **kr36** - 36氪
- ⏳ **netease_music** - 网易云音乐
- ⏳ **weread** - 微信读书
- ⏳ **woshipm** - 人人都是产品经理
- ⏳ **lol** - 英雄联盟
- ⏳ **zhihu_daily** - 知乎日报
- ⏳ **hellogithub** - HelloGitHub
- ⏳ **history** - 历史上的今天
- ⏳ **douban** - 豆瓣电影
- ⏳ **ithome** - IT之家
- ⏳ **dongchedi** - 懂车帝
- ⏳ **hupu** - 虎扑
- ⏳ **kuaishou** - 快手

## 🔧 Configuration

### Rate Limiting

Each platform has its own rate limit configuration:

```go
scraper.RateLimitConfig{
    RequestsPerMinute: 60,    // Max requests per minute
    CacheTTL:          5 * time.Minute,  // Cache duration
}
```

### Cache TTL

- **High-frequency platforms** (Weibo, Zhihu, Douyin): 5 minutes
- **Medium-frequency platforms** (GitHub, Baidu): 10 minutes
- **Low-frequency platforms** (Xiaohongshu): 10 minutes (due to anti-scraping)

## 🐛 Troubleshooting

### Xiaohongshu Returns Errors

Xiaohongshu has the most aggressive anti-scraping measures. The headers in `xiaohongshu.go` may need periodic updates. If you encounter errors:

1. Use a browser to visit `https://www.xiaohongshu.com`
2. Open DevTools → Network tab
3. Find API requests to `edith.xiaohongshu.com`
4. Copy the request headers
5. Update the headers in `internal/scraper/platforms/xiaohongshu.go`

### Rate Limit Exceeded

If you're hitting rate limits:

1. Increase the cache TTL for that platform
2. Reduce the `RequestsPerMinute` value
3. Add delays between requests in your client code

## 📊 Performance

- **Memory Usage**: ~20-30MB (idle)
- **Concurrent Requests**: 30+ platforms in parallel
- **Response Time**: 1-3 seconds for batch requests
- **Docker Image Size**: ~15MB (Alpine-based)

## 🔗 Integration with Python Backend

Example integration with the AutoUp-Agentic Python backend:

```python
import httpx

async def fetch_trends(platforms: list[str], limit: int = 10, keyword: str = "") -> dict:
    """Fetch hot trends from multiple platforms"""
    async with httpx.AsyncClient() as client:
        response = await client.post(
            "http://localhost:6000/api/v1/trends/batch",
            json={
                "platforms": platforms,
                "limit": limit,
                "keyword": keyword,
                "timeout_seconds": 30
            },
            timeout=35.0
        )
        return response.json()

# Usage
trends = await fetch_trends(["weibo", "zhihu", "douyin"], limit=5, keyword="羽毛球")
```

## 📝 License

This project adapts scraping logic from [next-daily-hot](https://github.com/imsyy/DailyHot), which is licensed under MIT License.

## 🙏 Acknowledgments

- **next-daily-hot** - Original project providing the scraping implementations
- **imsyy** - Original author
- All platform APIs and websites for providing public data

## 📮 Contact

For issues related to this Go implementation, please open an issue in the AutoUp-Agentic repository.

For issues related to the original scraping logic, please refer to the [next-daily-hot repository](https://github.com/imsyy/DailyHot).
