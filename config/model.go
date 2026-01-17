package config

type RabbitMQConfig struct {
	BrokerLink   string `mapstructure:"broker_link"`
	ExchangeName string `mapstructure:"exchange_name"`
	ExchangeType string `mapstructure:"exchange_type"`
	QueueName    string `mapstructure:"queue_name"`
	RoutingKey   string `mapstructure:"routing_key"`
	WorkerCount  int    `mapstructure:"worker_count"`
}
