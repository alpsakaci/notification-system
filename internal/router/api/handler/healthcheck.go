package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health godoc
// @Summary      System health check
// @Description  Returns the health status of the system.
// @Tags         system
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
