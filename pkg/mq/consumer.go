package mq

// import (
// 	"context"
// 	"errors"
// 	"log"
// 	"sync"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/rabbitmq/amqp091-go"
// )

// type Consumer struct {
// 	ch          *amqp091.Channel
// 	queueName   string
// 	workers     int
// 	sem         chan struct{}
// 	wg          sync.WaitGroup
// 	consumerTag string
// }

// func NewConsumer(conn *amqp091.Connection, queueName string, workers int) (*Consumer, error) {
// 	if conn == nil {
// 		return nil, errors.New("amqp connection is nil")
// 	}
// 	ch, err := conn.Channel()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := ch.Qos(workers, 0, false); err != nil {
// 		return nil, err
// 	}

// 	return &Consumer{
// 		ch:        ch,
// 		queueName: queueName,
// 		workers:   workers,
// 		sem:       make(chan struct{}, workers),
// 	}, nil
// }

// func (c *Consumer) Consumer(ctx context.Context, handler *EventHandler) error {
// 	c.consumerTag = uuid.NewString()

// 	msgs, err := c.ch.Consume(
// 		c.queueName,
// 		c.consumerTag,
// 		false,
// 		false,
// 		false,
// 		false,
// 		nil,
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	go func() {
// 		<-ctx.Done()
// 		_ = c.ch.Cancel(c.consumerTag, false)
// 	}()

// 	for msg := range msgs {
// 		c.sem <- struct{}{}
// 		c.wg.Add(1)

// 		go func(m amqp091.Delivery) {
// 			defer c.wg.Done()
// 			defer func() { <-c.sem }()

// 			msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 			defer cancel()

// 			if err := handler.Handle(msgCtx, m); err != nil {
// 				log.Printf("message failed: %v", err)
// 				_ = m.Nack(false, false)
// 				return
// 			}

// 			_ = m.Ack(false)

// 		}(msg)

// 	}

// 	c.wg.Wait()
// 	return nil
// }

// func (c *Consumer) Shutdown(ctx context.Context) error {
// 	if c.consumerTag != "" {
// 		_ = c.ch.Cancel(c.consumerTag, false)
// 	}

// 	done := make(chan struct{})

// 	go func() {
// 		c.wg.Wait()
// 		close(done)
// 	}()

// 	select {
// 	case <-done:
// 		return c.ch.Close()
// 	case <-ctx.Done():
// 		return ctx.Err()
// 	}
// }
