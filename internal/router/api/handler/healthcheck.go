package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"notification-system/internal/infrastructure/cache"
)

type HealthHandler struct {
	db          *gorm.DB
	redisClient *cache.RedisClient
}

func NewHealthHandler(db *gorm.DB, redisClient *cache.RedisClient) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisClient: redisClient,
	}
}

// Health godoc
// @Summary      Health Check
// @Description  Check the health of the application and its dependencies
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /api/v1/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	status := "healthy"
	details := make(map[string]string)

	// Check DB
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err == nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(ctx); err != nil {
				status = "unhealthy"
				details["database"] = "down"
			} else {
				details["database"] = "up"
			}
		}
	} else {
		details["database"] = "not_configured"
	}

	if h.redisClient != nil {
		details["redis"] = "up"
	} else {
		details["redis"] = "not_configured"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  status,
		"details": details,
	})
}
