package app

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartConsumer(ctx context.Context, c *Container) {
	go func() {
		log.Println("starting backgroud consumer")

		err := c.consumer.Consumer(ctx, &messageHandler{})
		if err != nil {
			log.Printf("Consumer stooped with error : %v", err)
		}
	}()
}

type messageHandler struct{}

func (h *messageHandler) Handle(ctx context.Context, m amqp.Delivery) error {
	log.Printf("Received message: %s", string(m.Body))
	return nil
}
