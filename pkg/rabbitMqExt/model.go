package rabbitMqExt

type RabbitMqExchangeConfig struct {
	Exchange   string
	QueueName  string
	RoutingKey string
	Channel    string

	DLQExchange   string
	DLQQueueName  string
	DLQRoutingKey string
}
