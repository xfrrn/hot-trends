package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DocsHandler provides OpenAPI schema and Swagger UI.
type DocsHandler struct{}

// NewDocsHandler creates a new docs handler.
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// SwaggerUI serves an interactive API docs page.
func (h *DocsHandler) SwaggerUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerHTML)
}

// OpenAPI serves OpenAPI 3.0 schema as JSON.
func (h *DocsHandler) OpenAPI(c *gin.Context) {
	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}

	serverURL := scheme + "://" + c.Request.Host

	spec := gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "Hot Trends API",
			"version":     "1.0.0",
			"description": "Interactive API documentation for hot-trends service.",
		},
		"servers": []gin.H{
			{"url": serverURL},
		},
		"paths": gin.H{
			"/health": gin.H{
				"get": gin.H{
					"summary":     "Health check",
					"operationId": "healthCheck",
					"responses": gin.H{
						"200": gin.H{
							"description": "OK",
							"content": gin.H{
								"application/json": gin.H{
									"schema": gin.H{"$ref": "#/components/schemas/HealthResponse"},
								},
							},
						},
					},
				},
			},
			"/api/v1/platforms": gin.H{
				"get": gin.H{
					"summary":     "List available platforms",
					"operationId": "listPlatforms",
					"responses": gin.H{
						"200": gin.H{
							"description": "Success",
							"content": gin.H{
								"application/json": gin.H{
									"schema": gin.H{
										"type": "object",
										"properties": gin.H{
											"platforms": gin.H{
												"type":  "array",
												"items": gin.H{"$ref": "#/components/schemas/PlatformInfo"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/trends/{platform}": gin.H{
				"get": gin.H{
					"summary":     "Get trends from one platform",
					"operationId": "getSinglePlatformTrends",
					"parameters": []gin.H{
						{
							"name":        "platform",
							"in":          "path",
							"required":    true,
							"description": "Platform identifier, e.g. weibo/zhihu/douyin",
							"schema":      gin.H{"type": "string"},
						},
						{
							"name":        "limit",
							"in":          "query",
							"required":    false,
							"description": "Number of items to return",
							"schema": gin.H{
								"type":    "integer",
								"minimum": 1,
								"maximum": 100,
								"default": 10,
							},
						},
						{
							"name":        "keyword",
							"in":          "query",
							"required":    false,
							"description": "Keyword filter",
							"schema":      gin.H{"type": "string"},
						},
					},
					"responses": gin.H{
						"200": gin.H{
							"description": "Success",
							"content": gin.H{
								"application/json": gin.H{
									"schema": gin.H{"$ref": "#/components/schemas/PlatformTrends"},
								},
							},
						},
					},
				},
			},
			"/api/v1/trends/batch": gin.H{
				"post": gin.H{
					"summary":     "Get trends from multiple platforms",
					"operationId": "getBatchTrends",
					"requestBody": gin.H{
						"required": true,
						"content": gin.H{
							"application/json": gin.H{
								"schema": gin.H{"$ref": "#/components/schemas/BatchRequest"},
							},
						},
					},
					"responses": gin.H{
						"200": gin.H{
							"description": "Success",
							"content": gin.H{
								"application/json": gin.H{
									"schema": gin.H{"$ref": "#/components/schemas/BatchResponse"},
								},
							},
						},
						"400": gin.H{
							"description": "Bad request",
							"content": gin.H{
								"application/json": gin.H{
									"schema": gin.H{"$ref": "#/components/schemas/ErrorResponse"},
								},
							},
						},
					},
				},
			},
		},
		"components": gin.H{
			"schemas": gin.H{
				"TrendItem": gin.H{
					"type": "object",
					"properties": gin.H{
						"id":         gin.H{"type": "string"},
						"title":      gin.H{"type": "string"},
						"desc":       gin.H{"type": "string"},
						"hot_value":  gin.H{"type": "integer", "format": "int64"},
						"url":        gin.H{"type": "string"},
						"mobile_url": gin.H{"type": "string"},
						"pic":        gin.H{"type": "string"},
						"label":      gin.H{"type": "string"},
					},
					"required": []string{"id", "title", "url"},
				},
				"PlatformTrends": gin.H{
					"type": "object",
					"properties": gin.H{
						"platform":   gin.H{"type": "string"},
						"items":      gin.H{"type": "array", "items": gin.H{"$ref": "#/components/schemas/TrendItem"}},
						"fetched_at": gin.H{"type": "string", "format": "date-time"},
						"cached":     gin.H{"type": "boolean"},
						"error":      gin.H{"type": "string"},
					},
					"required": []string{"platform", "items", "fetched_at", "cached"},
				},
				"BatchRequest": gin.H{
					"type": "object",
					"properties": gin.H{
						"platforms":       gin.H{"type": "array", "items": gin.H{"type": "string"}},
						"limit":           gin.H{"type": "integer", "default": 10},
						"keyword":         gin.H{"type": "string"},
						"timeout_seconds": gin.H{"type": "integer", "default": 30},
					},
					"required": []string{"platforms"},
				},
				"BatchResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"results":         gin.H{"type": "array", "items": gin.H{"$ref": "#/components/schemas/PlatformTrends"}},
						"total_platforms": gin.H{"type": "integer"},
						"successful":      gin.H{"type": "integer"},
						"failed":          gin.H{"type": "integer"},
					},
					"required": []string{"results", "total_platforms", "successful", "failed"},
				},
				"PlatformInfo": gin.H{
					"type": "object",
					"properties": gin.H{
						"id":                       gin.H{"type": "string"},
						"name":                     gin.H{"type": "string"},
						"type":                     gin.H{"type": "string"},
						"rate_limit":               gin.H{"type": "string"},
						"cache_ttl":                gin.H{"type": "string"},
						"enabled":                  gin.H{"type": "boolean"},
						"requires_special_headers": gin.H{"type": "boolean"},
					},
					"required": []string{"id", "name", "type", "rate_limit", "cache_ttl", "enabled"},
				},
				"HealthResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"status":              gin.H{"type": "string"},
						"timestamp":           gin.H{"type": "string", "format": "date-time"},
						"redis_connected":     gin.H{"type": "boolean"},
						"scrapers_registered": gin.H{"type": "integer"},
					},
					"required": []string{"status", "timestamp", "redis_connected", "scrapers_registered"},
				},
				"ErrorResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"error": gin.H{"type": "string"},
					},
					"required": []string{"error"},
				},
			},
		},
	}

	c.JSON(http.StatusOK, spec)
}

const swaggerHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Hot Trends API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; background: #f7f8fa; }
    #swagger-ui { max-width: 1100px; margin: 0 auto; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: window.location.origin + '/openapi.json',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      layout: 'BaseLayout'
    });
  </script>
</body>
</html>`
