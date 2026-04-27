package handlers

import (
	"hot-trends-service/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	redisConnected     func() bool
	scrapersRegistered int
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(redisConnected func() bool, scrapersRegistered int) *HealthHandler {
	return &HealthHandler{
		redisConnected:     redisConnected,
		scrapersRegistered: scrapersRegistered,
	}
}

// Handle handles the health check request
func (h *HealthHandler) Handle(c *gin.Context) {
	redisOK := true
	if h.redisConnected != nil {
		redisOK = h.redisConnected()
	}

	c.JSON(http.StatusOK, models.HealthResponse{
		Status:             "ok",
		Timestamp:          time.Now(),
		RedisConnected:     redisOK,
		ScrapersRegistered: h.scrapersRegistered,
	})
}
