package config

import (
	"os"
	"strconv"
)

type Config struct {
	RabbitURL     string
	QueueName     string
	WorkerCount   int
	PrefetchCount int
	Port          string
	RedistURL     string
}

func LoadEnv() *Config {
	return &Config{
		RabbitURL:     getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:     getenv("QUEUE_NAME", "scrape.jobs"),
		WorkerCount:   getenvInt("WORKER_COUNT", 5),
		PrefetchCount: getenvInt("PREFETCH_COUNT", 5),
		Port:          getenv("PORT", "8080"),
		RedistURL:     getenv("REDIS_URL", "redis://localhost:6379/0"),
	}
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}

	return fallback
}
