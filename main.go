package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "notification-system/docs"
	"notification-system/internal/router/api/handler"
)

// @title           Notification System API
// @version         1.0
// @description     This is a sample server for a notification system.
// @host            localhost:8080
// @BasePath        /

func main() {
	r := gin.Default()

	// Redirect root to swagger
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes group
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", handler.Health)
	}

	log.Println("Server is running at http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
