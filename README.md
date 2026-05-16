# Notification System 🚀

This is a high-performance, resilient, and scalable Notification System built with Go (Golang), Kafka, PostgreSQL, and Redis. It is designed to handle high volumes of notifications (Email, SMS, Push) with advanced features like rate limiting, idempotency, and robust dead-letter/retry mechanisms.

## 🏗 Architecture & Technologies

- **Golang**: Core application logic for both the API Server and Background Consumers.
- **PostgreSQL**: Persistent storage for notification records and their statuses.
- **Redis**: In-memory data store used for API Rate Limiting (100 req/sec) and Idempotency keys (preventing duplicate processing).
- **Apache Kafka**: Message broker handling four distinct topics based on priority and retries:
  - `notifications.high` (10 partitions)
  - `notifications.normal` (5 partitions)
  - `notifications.low` (1 partition)
  - `notifications.retry` (1 partition - Dead Letter Queue mechanism)
- **Docker & Docker Compose**: Containerization for seamless deployment and orchestration of all services.

## 🧠 Architecture Decisions & Technology Stack Justification

The design of this notification system prioritizes **throughput, reliability, and fault tolerance**. Below is the rationale behind each architectural choice:

### 1. Go (Golang)
- **Why we chose it**: Go's lightweight goroutines and excellent concurrency model make it the perfect candidate for high-throughput I/O bound applications (like reading from Kafka and making HTTP requests to external webhooks). It compiles to a single binary, resulting in minimal memory footprint and fast startup times.

### 2. Apache Kafka
- **Why we chose it**: Rather than using a simpler message queue like RabbitMQ or Redis Pub/Sub, Kafka was selected because of its **durability** and **partitioning** capabilities.
- **Partition Strategy**: By assigning 10 partitions to the `high` priority topic, we can effortlessly scale up to 10 concurrent consumers for critical notifications. Kafka's persistent log ensures that if all consumers crash, no messages are lost.

### 3. Redis
- **Why we chose it**: Redis provides lightning-fast read/write operations for ephemeral data.
- **Use Cases**:
  - **Idempotency**: Preventing duplicate notification deliveries requires a distributed lock/cache mechanism. Redis `SETNX` handles this elegantly.
  - **Rate Limiting**: To protect our API from DDOS or abusive clients, Redis is used to track the number of requests per IP address in real-time.

### 4. PostgreSQL
- **Why we chose it**: While Kafka acts as the event stream, we need a reliable source of truth to track the lifecycle of a notification (Pending ➔ Processing ➔ Sent ➔ Failed). PostgreSQL provides robust ACID transactions and relational querying capabilities, allowing us to easily build dashboard endpoints (e.g., "List all failed SMS notifications within the last 24 hours").

### 5. Dead Letter Queue (DLQ) & Retry Strategy
- **Why we chose it**: External providers (Email servers, SMS APIs) often experience transient failures or rate limits. Instead of blocking the main processing pipeline or dropping the message entirely, the system uses an **Exponential Backoff** strategy in-memory. If the external provider remains unavailable, the message is safely offloaded to the `notifications.retry` topic (our DLQ implementation). A dedicated consumer group can then process this retry topic at a slower, controlled pace without impacting high-priority traffic.


## ✨ Key Features

1. **Priority-Based Routing**: Notifications are routed to different Kafka topics based on their priority (high, normal, low) to ensure critical messages are processed immediately.
2. **Robust Retry Mechanism (DLQ)**: If a webhook delivery fails, the system implements an Exponential Backoff retry strategy (e.g., 2s, 4s, 8s). If it fails 3 times, it is pushed to the `notifications.retry` topic to be reprocessed safely later.
3. **API-Level Rate Limiting**: Redis-backed middleware protects the API from abuse, limiting incoming requests to 100 per second per IP (`HTTP 429 Too Many Requests`).
4. **Idempotency Check**: Redis ensures that if a notification is accidentally delivered twice by Kafka, the Consumer ignores the duplicate.
5. **Observability**: Exposes Prometheus metrics (e.g., processed messages, rate limit hits) on both API and Consumer layers.
6. **Graceful Degradation**: The services have retry logic to withstand database or broker connection drops during startup.

## 🚀 Getting Started

### Prerequisites
- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed on your machine.

### Installation & Running

1. Clone the repository and navigate to the root directory.
2. Build and start all services in detached mode:
   ```bash
   docker-compose up -d --build
   ```

This command will spin up the following containers:
- `notification_api` (localhost:8080)
- `notification_consumer` (localhost:8082)
- `notification_postgres` (localhost:5432)
- `notification_redis` (localhost:6379)
- `notification_kafka` (localhost:9092)
- `notification_kafka_ui` (localhost:8081) - Web UI for Kafka
- `notification_pgadmin` (localhost:5050) - Web UI for PostgreSQL

### Monitoring Interfaces
- **Swagger API Docs**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- **Kafka UI**: [http://localhost:8081](http://localhost:8081)
- **pgAdmin**: [http://localhost:5050](http://localhost:5050) (admin@admin.com / admin)
- **API Metrics (Prometheus)**: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- **Consumer Metrics (Prometheus)**: [http://localhost:8082/metrics](http://localhost:8082/metrics)
- **Webhook Inspector (Live Logs)**: [https://webhook.site/#!/view/6762e0e1-47a2-492d-855e-fb974baca2b6/23b59863-146f-4294-b6cf-76fa2de34e52/1](https://webhook.site/#!/view/6762e0e1-47a2-492d-855e-fb974baca2b6/23b59863-146f-4294-b6cf-76fa2de34e52/1)

## 📡 API Endpoints

### 1. Create Notification
`POST /api/v1/notifications`
```json
{
  "recipient": "user@example.com",
  "channel": "email",
  "content": "Hello World!",
  "priority": "high"
}
```

### 2. Batch Create Notifications
`POST /api/v1/notifications/batch`
*(Max 1000 items per request)*
```json
{
  "notifications": [
    {
      "recipient": "user1@example.com",
      "channel": "email",
      "content": "Test 1",
      "priority": "normal"
    }
  ]
}
```

### 3. Check Status
`GET /api/v1/notifications/:id`

### 4. Cancel Notification
`PUT /api/v1/notifications/:id/cancel`

### 5. List Notifications
`GET /api/v1/notifications`

### 6. Health Check
`GET /api/v1/health`

## 🛠 Testing Rate Limiting
A bash script (`test_rate_limit.sh`) is included to test the API rate limiting by firing 500 parallel asynchronous requests to the API.
```bash
chmod +x test_rate_limit.sh
./test_rate_limit.sh
```
You will see 100 requests succeed (`HTTP 202`), and the rest will be blocked (`HTTP 429`).

## 📜 Environment Variables
- `WEBHOOK_URL`: The destination URL where the Consumer will POST the notification payloads. Edit `docker-compose.yml` to set your actual webhook destination (e.g., webhook.site).

## 🌐 Live Server / Deployment
This project is currently deployed and running live on the following server: **`37.148.213.87`**

> ⚠️ **Note on Server Capacity:** This live demo is hosted on a small, low-tier server. Due to hardware constraints, the system may occasionally operate slowly or experience crashes if subjected to heavy load testing or extreme rate limit spikes. Please be gentle! 😅

You can access the live monitoring and API interfaces using the links below:
- **Swagger API Docs**: [http://37.148.213.87:8080/swagger/index.html](http://37.148.213.87:8080/swagger/index.html)
- **Kafka UI**: [http://37.148.213.87:8081](http://37.148.213.87:8081)
- **pgAdmin**: [http://37.148.213.87:5050](http://37.148.213.87:5050)
- **API Metrics**: [http://37.148.213.87:8080/metrics](http://37.148.213.87:8080/metrics)
- **Consumer Metrics**: [http://37.148.213.87:8082/metrics](http://37.148.213.87:8082/metrics)
- **Webhook Inspector (Live Logs)**: [https://webhook.site/#!/view/6762e0e1-47a2-492d-855e-fb974baca2b6/23b59863-146f-4294-b6cf-76fa2de34e52/1](https://webhook.site/#!/view/6762e0e1-47a2-492d-855e-fb974baca2b6/23b59863-146f-4294-b6cf-76fa2de34e52/1)
