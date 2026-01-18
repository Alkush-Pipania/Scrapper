package config

import (
	"os"
	"strconv"
)

type Config struct {
	RabbitMQ        RabbitMQConfig
	PrefetchCount   int
	Port            string
	RedisURL        string
	Env             string
	TurnstileSecret string
}

func LoadEnv() *Config {
	return &Config{
		RabbitMQ: RabbitMQConfig{
			BrokerLink:   getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			ExchangeName: getenv("EXCHANGE_NAME", "scrape.exchange"),
			ExchangeType: getenv("EXCHANGE_TYPE", "direct"),
			QueueName:    getenv("QUEUE_NAME", "scrape.jobs"),
			RoutingKey:   getenv("ROUTING_KEY", "scrape"),
			WorkerCount:  getenvInt("WORKER_COUNT", 5),
		},
		PrefetchCount:   getenvInt("PREFETCH_COUNT", 5),
		Port:            getenv("PORT", "8080"),
		RedisURL:        getenv("REDIS_URL", "localhost:6379"),
		Env:             getenv("ENV", "development"),
		TurnstileSecret: getenv("TURNSTILE_SECRET_KEY", ""),
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
