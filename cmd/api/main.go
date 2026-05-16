package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "notification-system/docs"
	"notification-system/internal/application/command"
	"notification-system/internal/application/query"
	"notification-system/internal/infrastructure/cache"
	"notification-system/internal/infrastructure/database"
	"notification-system/internal/infrastructure/messaging"
	"notification-system/internal/router/api/handler"
	"notification-system/internal/router/middleware"
)

// @title           Notification System API
// @version         1.0
// @description     This is a scalable server for a notification system.
// @BasePath        /

func main() {
	// Initialize Structured Logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Initialize Database (Use environment variables in a real app)
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dsn := "host=" + dbHost + " user=postgres password=postgres dbname=notification_db port=5432 sslmode=disable TimeZone=UTC"
	// Change host to postgres if running inside docker-compose
	db, err := database.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	} else {
		// Run migrations
		if err := database.Migrate(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	// Initialize Redis
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost:6379"
	}
	redisClient, err := cache.NewRedisClient(redisHost)
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	}

	// Initialize Kafka Producer
	kafkaHost := os.Getenv("KAFKA_HOST")
	if kafkaHost == "" {
		kafkaHost = "localhost:9092"
	}
	brokers := []string{kafkaHost}

	// Initialize Topics
	messaging.InitTopics(brokers)

	producer := messaging.NewKafkaProducer(brokers)
	// Change host to kafka if running inside docker-compose
	defer producer.Close()

	// Initialize Repositories
	repo := database.NewNotificationRepository(db)

	// Initialize Application Logic Handlers
	createCmd := command.NewCreateNotificationHandler(repo, producer)
	batchCmd := command.NewBatchCreateNotificationHandler(repo, producer)
	cancelCmd := command.NewCancelNotificationHandler(repo)
	getQry := query.NewGetNotificationHandler(repo)
	listQry := query.NewListNotificationsHandler(repo)

	// Initialize HTTP Handlers
	notiHandler := handler.NewNotificationHandler(createCmd, cancelCmd, getQry, listQry, batchCmd)
	healthHandler := handler.NewHealthHandler(db, redisClient)

	r := gin.Default()

	// Redirect root to swagger
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Metrics route
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes group
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Health)

		// Notification routes - Apply Rate Limiting here (e.g. max 100 requests per second per IP)
		notifGroup := v1.Group("/notifications")
		notifGroup.Use(middleware.RateLimitMiddleware(redisClient, 100))
		{
			notifGroup.POST("", notiHandler.Create)
			notifGroup.POST("/batch", notiHandler.BatchCreate)
			notifGroup.GET("/:id", notiHandler.Get)
			notifGroup.PUT("/:id/cancel", notiHandler.Cancel)
			notifGroup.GET("", notiHandler.List)
		}
	}

	log.Println("Server is running at http://0.0.0.0:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
