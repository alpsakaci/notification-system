package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "notification-system/docs"
)

// @title           Notification System API
// @version         1.0
// @description     This is a sample server for a notification system.
// @host            localhost:8080
// @BasePath        /api/v1

func main() {
	r := gin.Default()

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes group
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", Ping)
	}

	log.Println("Server is running at http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

// Ping godoc
// @Summary      Ping the server
// @Description  Responds with a simple message to check if the server is up and running.
// @Tags         system
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
