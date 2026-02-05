# Go Scraper Service

An async web‑scraping service built with Go, RabbitMQ, and Redis. It accepts a URL, queues a job, scrapes in the background, and lets clients poll for status/results. It supports static HTML parsing, Browserless (headless) fallback, and YouTube metadata scraping.

---

**Highlights**
1. Async job queue with RabbitMQ + Redis status tracking
2. Static scrape first, Browserless fallback when needed
3. Screenshot capture (uploaded to S3/Spaces) when OG image is missing
4. YouTube URLs use the YouTube Data API + transcript fetch

---

**Architecture (high level)**
1. Client `POST /api/v1/scrape` → job saved to Redis → message published to RabbitMQ
2. Worker consumes job → runs scraper engine → updates Redis with result or error
3. Client `GET /api/v1/scrape/{id}` polls status

---

**Project Structure**
- `internal/modules/scrape` → HTTP handler, service, worker
- `internal/modules/scrape/engine` → core scraping logic (static + browser + YouTube)
- `internal/infra/*` → adapters for Browserless and YouTube
- `pkg/*` → external integrations (RabbitMQ, Redis, S3, Browserless, YouTube)

---

## Getting Started

### 1) Prerequisites
- Go 1.21+
- Docker (for RabbitMQ + Redis)

### 2) Start dependencies
```bash
# Redis
docker run -d -p 6379:6379 --name redis redis

# RabbitMQ
docker run -d -p 5672:5672 -p 15672:15672 --name rabbitmq rabbitmq:3-management
```

### 3) Configure environment
Create `.env` in project root.

**Required (minimal to run locally)**
```env
PORT=8082
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
REDIS_URL=localhost:6379
EXCHANGE_NAME=scrape.exchange
EXCHANGE_TYPE=direct
QUEUE_NAME=scrape.jobs
ROUTING_KEY=scrape
WORKER_COUNT=5
```

**Optional (Browserless + S3 for screenshots)**
```env
# Browserless Cloud (BrowserQL)
BURL=https://production-sfo.browserless.io
BTOKEN=your_browserless_token

# DigitalOcean Spaces / S3
DO_REGION=blr1
DO_ENDPOINT=blr1.digitaloceanspaces.com
DO_ACCESS_KEY=your_key
DO_SECRET_KEY=your_secret
DO_BUCKET=your_bucket
```

**Optional (YouTube)**
```env
YOUTUBE_API_KEY=your_youtube_api_key
```

### 4) Run the API
```bash
go run cmd/api/main.go
```

---

## API Usage

### Submit a job
```http
POST /api/v1/scrape
Content-Type: application/json
```

```json
{
  "url": "https://example.com"
}
```

Response:
```json
{
  "job_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef"
}
```

### Poll status
```http
GET /api/v1/scrape/{job_id}
```

Processing:
```json
{
  "id": "...",
  "url": "https://example.com",
  "status": "processing"
}
```

Completed:
```json
{
  "id": "...",
  "url": "https://example.com",
  "status": "completed",
  "result": {
    "title": "Example Domain",
    "description": "...",
    "image_url": "https://.../screenshots/123.jpg",
    "site_name": "Example",
    "content_text": "..."
  }
}
```

---

## Scrape Strategy

1. **Static scrape** (go-readability + goquery) runs first.
2. If static result is weak or missing image → **Browserless** fallback.
3. If Browserless returns a screenshot, it is uploaded to S3/Spaces and returned as `image_url`.
4. YouTube URLs bypass the above and use the YouTube API.

---

## Notes

- Redis job TTL is **24 hours** (see `pkg/redis/client.go`).
- Screenshots in S3/Spaces are **not automatically deleted** in code. To match Redis TTL, set a Space lifecycle rule for the `screenshots/` prefix (expire after 1 day).
- Browserless uses **BrowserQL** at `BURL/chromium/bql?token=...`.
- If `go test ./...` fails due to Go cache permissions, run:
  ```bash
  GOCACHE=/tmp/go-build go test ./...
  ```

---

## Security

- Never commit `.env` to GitHub.
- Rotate any leaked keys immediately.
- Consider using `.env.example` for safe sharing.

---

## License

MIT (or update this section if you use another license)
