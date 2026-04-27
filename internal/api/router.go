package api

import (
	"hot-trends-service/internal/api/handlers"
	"hot-trends-service/internal/scraper"

	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the HTTP router
func SetupRouter(executor *scraper.Executor, registry *scraper.Registry, redisConnected func() bool) *gin.Engine {
	r := gin.Default()

	// Interactive API docs
	docsHandler := handlers.NewDocsHandler()
	r.GET("/docs", docsHandler.SwaggerUI)
	r.GET("/openapi.json", docsHandler.OpenAPI)

	// Health check
	healthHandler := handlers.NewHealthHandler(redisConnected, registry.Count())
	r.GET("/health", healthHandler.Handle)

	// API v1
	v1 := r.Group("/api/v1")
	{
		trendsHandler := handlers.NewTrendsHandler(executor, registry)

		// Trends endpoints
		v1.GET("/trends/:platform", trendsHandler.GetSingle)
		v1.POST("/trends/batch", trendsHandler.GetBatch)

		// Platforms endpoints
		v1.GET("/platforms", trendsHandler.ListPlatforms)
	}

	return r
}
