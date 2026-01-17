package mq

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch         *amqp.Channel            // AMQP channel for publishing messages
	confirms   <-chan amqp.Confirmation // Channel to receive publish confirmations
	exchange   string                   // Exchange to publish messages to
	routingKey string                   // Routing key for the messages
}

func NewPublisher(conn *amqp.Connection, exchange string, routingkey string) (*Publisher, error) {
	if conn == nil {
		return nil, errors.New("AMQP connection is nil ")
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.Confirm(false); err != nil {
		return nil, err
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 100))

	return &Publisher{
		ch:         ch,
		confirms:   confirms,
		exchange:   exchange,
		routingKey: routingkey,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, body []byte) error {
	if p.ch == nil {
		return errors.New("AMQP chanel is nil ")
	}
	return p.ch.PublishWithContext(
		ctx,
		p.exchange,
		p.routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Publisher) Close() error {
	if p.ch != nil {
		return p.ch.Close()
	}
	return nil
}
