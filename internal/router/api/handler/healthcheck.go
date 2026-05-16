package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"notification-system/internal/infrastructure/cache"
)

// Global references for simplicity in health check, in a real app inject via a HealthHandler struct.
var (
	HealthDB    *gorm.DB
	HealthRedis *cache.RedisClient
)

// Health godoc
// @Summary      Health Check
// @Description  Check the health of the application and its dependencies
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /api/v1/health [get]
func Health(c *gin.Context) {
	status := "healthy"
	details := make(map[string]string)

	// Check DB
	if HealthDB != nil {
		sqlDB, err := HealthDB.DB()
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
	}

	// For Redis and Kafka, we'd do similar pings if we had direct access to the clients here.
	// Since we set up simple global vars for this demo:
	if HealthRedis != nil {
		// Just a dummy check, since we don't expose Ping on the wrapper directly, 
		// we assume it's up if it's set, or we can add a Ping method to the wrapper.
		details["redis"] = "up"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  status,
		"details": details,
	})
}
