package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Consumer struct {
	ch        *amqp.Channel
	exchange  string
	queueName string
	routes    []string
	name      string
}

func NewConsumer(channel *amqp.Channel, exchangeName, consumerName, queueName string, routingKeys []string) *Consumer {
	return &Consumer{
		ch:        channel,
		exchange:  exchangeName,
		queueName: queueName,
		routes:    routingKeys,
		name:      consumerName,
	}
}

func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	for _, route := range c.routes {
		err := c.ch.QueueBind(c.queueName, route, c.exchange, false, nil)
		if err != nil {
			log.Printf("queue bind err: %s. queueName = %s, route = %s, exchange = %s\n",
				err, c.queueName, route, c.exchange)
		}
	}

	consume, err := c.ch.Consume(c.queueName, c.name, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("channel.Consume() err: %w", err)
	}
	return consume, nil
}
