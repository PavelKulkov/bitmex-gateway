package rabbitmq

import (
	"encoding/json"
	"fmt"

	"provider/domain/models"

	"github.com/streadway/amqp"
)

type Producer struct {
	ch       *amqp.Channel
	exchange string
}

func NewProducer(exchangeName string, channel *amqp.Channel) *Producer {
	return &Producer{
		ch:       channel,
		exchange: exchangeName,
	}
}

func (p *Producer) Publish(data models.Data) error {
	dataJson, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json marshal message err: %w", err)
	}
	publishing := amqp.Publishing{
		ContentType: "application/json",
		Body:        dataJson,
	}
	err = p.ch.Publish(p.exchange, data.Symbol, false, false, publishing)
	if err != nil {
		return err
	}

	err = p.ch.Publish(p.exchange, "all", false, false, publishing)
	if err != nil {
		return err
	}
	return nil
}
