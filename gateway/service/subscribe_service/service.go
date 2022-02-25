package subscribe_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gateway/domain/models"
	"gateway/rabbitmq"

	"github.com/streadway/amqp"
)

type Service struct {
	ch       *amqp.Channel
	exchange string
}

func New(channel *amqp.Channel, exchangeName string) *Service {
	return &Service{
		ch:       channel,
		exchange: exchangeName,
	}
}

func (s *Service) Subscribe(ctx context.Context, clientName, queueName string, symbols []string, dataCh chan<- models.Data) error {
	queue, err := s.ch.QueueDeclare(queueName, false, true, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare rabbitmq queue err: %w", err)
	}
	consumer := rabbitmq.NewConsumer(s.ch, s.exchange, clientName, queue.Name, symbols)

	consume, err := consumer.Consume()
	if err != nil {
		return fmt.Errorf("consume err: %w", err)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				delivery, ok := <-consume
				if !ok {
					return
				}
				data := models.Data{}
				unmarshalErr := json.Unmarshal(delivery.Body, &data)
				if unmarshalErr != nil {
					log.Printf("delivered message json unmarshal err: %s\n", unmarshalErr)
					continue
				}
				dataCh <- data
			}
		}
	}()
	return nil
}

func (s *Service) Unsubscribe(queueName string) error {
	_, err := s.ch.QueueDelete(queueName, false, false, false)
	if err != nil {
		return fmt.Errorf("queue delete err: %s", err)
	}
	return nil
}
