package mq

// import (
// 	"errors"
// 	"log"
// 	"time"

// 	"github.com/Alkush-Pipania/Scrapper/config"
// 	"github.com/rabbitmq/amqp091-go"
// )

// func NewConn(rmqCfg *config.RabbitMQConfig) (*amqp091.Connection, error) {
// 	var conn *amqp091.Connection
// 	var err error

// 	for i := range 5 {
// 		conn, err = amqp091.Dial(rmqCfg.BrokerLink)
// 		if err == nil {
// 			return conn, nil
// 		}
// 		time.Sleep(2 * time.Second)
// 		log.Printf("rabbitmq mq connection attempted %v", i+1)
// 	}
// 	log.Printf("failed to connect to rabbitmq, after %v attempts : %v ", 5, err)
// 	return nil, errors.New("failed to connect to rabbitmq")
// }

// func SetupTopology(conn *amqp091.Connection, rmqCfg *config.RabbitMQConfig) error {
// 	ch, err := conn.Channel()
// 	if err != nil {
// 		return err
// 	}
// 	defer ch.Close()

// 	if err := ch.ExchangeDeclare(
// 		rmqCfg.ExchangeName,
// 		rmqCfg.ExchangeType,
// 		true, false, false, false, nil,
// 	); err != nil {
// 		return err
// 	}

// 	if _, err := ch.QueueDeclare(
// 		rmqCfg.QueueName,
// 		true, false, false, false, nil,
// 	); err != nil {
// 		return err
// 	}

// 	if err = ch.QueueBind(
// 		rmqCfg.QueueName,
// 		rmqCfg.RoutingKey,
// 		rmqCfg.ExchangeName,
// 		false, nil,
// 	); err != nil {
// 		return err
// 	}
// 	return nil
// }
