# Go Scraper Service 🕷️

> A production-grade, asynchronous web scraping service built with Go, RabbitMQ, and Redis.

A scalable solution for extracting Open Graph (OG) data from websites without blocking the main API. Uses an event-driven architecture to handle scraping jobs in the background, making it robust enough for high concurrency and graceful failure handling.

---

## ️ Tech Stack

| Technology | Purpose |
|------------|---------|
| **Go** 1.21+ | Core language |
| **Chi** | Lightweight, idiomatic HTTP router |
| **RabbitMQ** | Message broker |
| **Redis** | State management & caching |

**Architecture:** Modular Monolith (Clean Architecture)

---

## 🏃 Getting Started

### Prerequisites

- **Go** 1.21+
- **Docker** (for RabbitMQ & Redis)

Start the dependencies:

```bash
# Start Redis
docker run -d -p 6379:6379 --name redis redis

# Start RabbitMQ
docker run -d -p 5672:5672 -p 15672:15672 --name rabbitmq rabbitmq:3-management
```

### Installation

```bash
# Clone the repository
git clone https://github.com/Alkush-Pipania/Scrapper.git
cd Scrapper

# Install dependencies
go mod download
```

### Configuration

Create a `.env` file in the root directory:

```env
PORT=8080
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
REDIS_URL=redis://localhost:6379/0
QUEUE_NAME=scrape.jobs
WORKER_COUNT=5
```

### Run

```bash
go run cmd/api/main.go
```

---

## 🔌 API Reference

### Submit a Job

```http
POST /scrape
Content-Type: application/json
```

**Request:**
```json
{
  "url": "https://github.com"
}
```

**Response:** `202 Accepted`
```json
{
  "job_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef"
}
```

---

### Check Status

```http
GET /scrape/{job_id}
```

**Response (Processing):**
```json
{
  "id": "a1b2c3d4...",
  "url": "https://github.com",
  "status": "processing"
}
```

**Response (Completed):**
```json
{
  "id": "a1b2c3d4...",
  "url": "https://github.com",
  "status": "completed",
  "result": {
    "title": "GitHub: Let's build from here",
    "description": "GitHub is where over 100 million developers shape the future of software...",
    "og_image": "https://github.githubassets.com/images/modules/site/social-cards/github-social.png"
  }
}
```

---
