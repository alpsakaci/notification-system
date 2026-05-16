package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"notification-system/internal/application/consumer"
	"notification-system/internal/infrastructure/cache"
	"notification-system/internal/infrastructure/database"
	"notification-system/internal/infrastructure/messaging"
)

func main() {
	// Initialize Structured Logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Notification Consumer starting...")

	// Initialize Database (Use environment variables in a real app)
	dsn := "host=localhost user=postgres password=postgres dbname=notification_db port=5432 sslmode=disable TimeZone=UTC"
	// Change host to postgres if running inside docker-compose
	db, err := database.NewPostgresDB(dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}
	repo := database.NewNotificationRepository(db)

	// Initialize Redis
	redisClient, err := cache.NewRedisClient("localhost:6379")
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	}

	// Read webhook URL from env
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		webhookURL = "https://webhook.site/placeholder-uuid" // Should be replaced with actual UUID provided by user
	}

	// Initialize Worker
	worker := consumer.NewWorker(repo, redisClient, webhookURL)

	// Initialize Kafka Consumers for the priority topics
	brokers := []string{"localhost:9092"}
	// Change to kafka:9092 if running inside docker-compose
	groupID := "notification-workers"
	topics := []string{"notifications.high", "notifications.normal", "notifications.low"}
	
	var consumers []*messaging.KafkaConsumer
	ctx, cancel := context.WithCancel(context.Background())

	for _, topic := range topics {
		c := messaging.NewKafkaConsumer(brokers, groupID, topic)
		consumers = append(consumers, c)
		
		go func(kafkaConsumer *messaging.KafkaConsumer, t string) {
			log.Printf("Listening to topic: %s", t)
			kafkaConsumer.Start(ctx, worker.ProcessMessage)
		}(c, topic)
	}

	// Start Metrics and Health Server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"status\": \"healthy\", \"service\": \"consumer\"}"))
		})
		slog.Info("Starting metrics server on :8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			slog.Error("Metrics server failed", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Notification Consumer shutting down...")
	cancel()

	for _, c := range consumers {
		_ = c.Close()
	}
	slog.Info("Shutdown complete")
}
