package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "notification-system/docs"
	"notification-system/internal/application/command"
	"notification-system/internal/application/query"
	"notification-system/internal/infrastructure/database"
	"notification-system/internal/infrastructure/messaging"
	"notification-system/internal/router/api/handler"
)

// @title           Notification System API
// @version         1.0
// @description     This is a scalable server for a notification system.
// @host            localhost:8080
// @BasePath        /

func main() {
	// Initialize Database (Use environment variables in a real app)
	dsn := "host=localhost user=postgres password=postgres dbname=notification_db port=5432 sslmode=disable TimeZone=UTC"
	// Change host to postgres if running inside docker-compose
	db, err := database.NewPostgresDB(dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		// We log instead of fatal for now so the app can start even if DB is down locally, though usually you want to crash here.
	} else {
		// Run migrations
		if err := database.Migrate(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	// Initialize Kafka Producer
	producer := messaging.NewKafkaProducer([]string{"localhost:9092"})
	// Change host to kafka if running inside docker-compose
	defer producer.Close()

	// Initialize Repositories
	repo := database.NewNotificationRepository(db)

	// Initialize Application Logic Handlers
	createCmd := command.NewCreateNotificationHandler(repo, producer)
	cancelCmd := command.NewCancelNotificationHandler(repo)
	getQry := query.NewGetNotificationHandler(repo)
	listQry := query.NewListNotificationsHandler(repo)

	// Initialize HTTP Handlers
	notiHandler := handler.NewNotificationHandler(createCmd, cancelCmd, getQry, listQry)

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

		// Notification routes
		v1.POST("/notifications", notiHandler.Create)
		v1.GET("/notifications/:id", notiHandler.Get)
		v1.PUT("/notifications/:id/cancel", notiHandler.Cancel)
		v1.GET("/notifications", notiHandler.List)
	}

	log.Println("Server is running at http://0.0.0.0:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
