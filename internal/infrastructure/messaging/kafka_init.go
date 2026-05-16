package messaging

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// InitTopics ensures that the necessary Kafka topics exist with the desired partition counts.
func InitTopics(brokers []string) {
	if len(brokers) == 0 {
		return
	}
	broker := brokers[0]

	// Add a small retry loop in case Kafka is not fully ready
	var conn *kafka.Conn
	var err error
	for i := 0; i < 5; i++ {
		conn, err = kafka.Dial("tcp", broker)
		if err == nil {
			break
		}
		log.Printf("Failed to dial kafka %s: %v, retrying...", broker, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Printf("Could not connect to Kafka for topic initialization: %v", err)
		return
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Printf("Could not get Kafka controller: %v", err)
		return
	}

	controllerAddr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	controllerConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		log.Printf("Could not dial Kafka controller: %v", err)
		return
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{Topic: "notifications.high", NumPartitions: 10, ReplicationFactor: 1},
		{Topic: "notifications.normal", NumPartitions: 5, ReplicationFactor: 1},
		{Topic: "notifications.low", NumPartitions: 1, ReplicationFactor: 1},
		{Topic: "notifications.retry", NumPartitions: 1, ReplicationFactor: 1},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Printf("CreateTopics returned an error (might be benign if topics already exist): %v", err)
	} else {
		log.Println("Kafka topics initialized successfully with specified partitions.")
	}
}
