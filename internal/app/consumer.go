package app

import (
	"context"
	"log"
)

func StartConsumer(ctx context.Context, c *Container) {
	go func() {
		log.Println("starting background consumer")

		err := c.consumer.Consumer(ctx, c.ScrapeWk)
		if err != nil {
			log.Printf("Consumer stooped with error : %v", err)
		}
	}()
}
