package app

import (
	"context"

	"github.com/Alkush-Pipania/Scrapper/config"
	"github.com/Alkush-Pipania/Scrapper/internal/modules/scrape"
	"github.com/Alkush-Pipania/Scrapper/pkg/mq"
	"github.com/Alkush-Pipania/Scrapper/pkg/redis"
	"github.com/Alkush-Pipania/Scrapper/pkg/turnstile"
	"github.com/rabbitmq/amqp091-go"
)

type Container struct {
	ScrapeHandler *scrape.Handler
	consumer      *mq.Consumer
	RMQConn       *amqp091.Connection
	ScrapeWk      *scrape.ScrapeWorker
}

func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	// setup rabbit mq connection and consumer
	rmqpConn, consumer, err := setupRabbitMQ(&cfg.RabbitMQ)
	if err != nil {
		return nil, err
	}
	// setup redis
	rds := redis.NewRedis(cfg.RedisURL)

	// setup rabbitmq publisher
	pbh, err := mq.NewPublisher(rmqpConn, cfg.RabbitMQ.ExchangeName, cfg.RabbitMQ.RoutingKey)
	if err != nil {
		return nil, err
	}

	tsClient := turnstile.New(cfg.TurnstileSecret)

	scrapeWorker := scrape.NewScrapeWorker(rds)
	scrapeService := scrape.NewService(rds, pbh)
	scrapeHandler := scrape.NewHandler(scrapeService, tsClient)
	return &Container{
		ScrapeHandler: scrapeHandler,
		consumer:      consumer,
		RMQConn:       rmqpConn,
		ScrapeWk:      scrapeWorker,
	}, nil
}

func setupRabbitMQ(cfg *config.RabbitMQConfig) (*amqp091.Connection, *mq.Consumer, error) {
	rmqpConn, err := mq.NewConn(cfg)
	if err != nil {
		return nil, nil, err
	}
	if err := mq.SetupTopology(rmqpConn, cfg); err != nil {
		return nil, nil, err
	}
	consumer, err := mq.NewConsumer(rmqpConn, cfg.QueueName, cfg.WorkerCount)
	if err != nil {
		return nil, nil, err
	}
	return rmqpConn, consumer, nil

}

func (c *Container) Shutdown(ctx context.Context) error {
	if c.consumer != nil {
		_ = c.consumer.Shutdown(ctx)
	}

	if c.RMQConn != nil {
		_ = c.RMQConn.Close()
	}

	return nil
}
